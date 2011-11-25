package config

import (
	"os"
	"io"
	"xml"
	"fmt"
	"strings"
)

type xmlNode struct {
	name       string
	attributes map[string]string
	children   []*xmlNode
	value      string
}

func newNode() *xmlNode {
	node := new(xmlNode)
	node.children = make([]*xmlNode, 0)
	node.attributes = make(map[string]string)
	return node
}

func (this *xmlNode) String() string {
	str := fmt.Sprintf("<%s", this.name)

	for attrName, attrVal := range this.attributes {
		str += fmt.Sprintf(" %s=\"%s\"", attrName, attrVal)
	}

	str += ">"
	str += this.value

	if len(this.children) != 0 {
		for _, child := range this.children {
			str += fmt.Sprintf("%s", child)
		}
	}

	str += fmt.Sprintf("</%s>", this.name)

	return str
}

func (this *xmlNode) unmarshal(startEl xml.StartElement) os.Error {
	this.name = startEl.Name.Local

	for _, v := range startEl.Attr {
		_, alreadyExists := this.attributes[v.Name.Local]
		if alreadyExists {
			return os.NewError("Tag '" + this.name + "' has duplicated attribute: '" + v.Name.Local + "'")
		}
		this.attributes[v.Name.Local] = v.Value
	}

	return nil
}

func (this *xmlNode) add(child *xmlNode) {
	if this.children == nil {
		this.children = make([]*xmlNode, 0)
	}

	this.children = append(this.children, child)
}

//=============================================

func unmarshalConfig(reader io.Reader) (*xmlNode, os.Error) {
	xmlParser := xml.NewParser(reader)

	config, err := unmarshalNode(xmlParser, nil)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, os.NewError("Xml has no content")
	}

	nextConfigEntry, err := unmarshalNode(xmlParser, nil)
	if nextConfigEntry != nil {
		return nil, os.NewError("Xml contains more than one root element")
	}

	return config, nil
}

func unmarshalNode(xmlParser *xml.Parser, curToken xml.Token) (node *xmlNode, err os.Error) {
	firstLoop := true
	for {
		var tok xml.Token
		if firstLoop && curToken != nil {
			tok = curToken
			firstLoop = false
		} else {
			tok, err = getNextToken(xmlParser)
			if err != nil || tok == nil {
				return
			}
		}

		switch tt := tok.(type) {
		case xml.SyntaxError:
			err = os.NewError(tt.String())
			return
		case xml.CharData:
			value := strings.TrimSpace(string([]byte(tt)))
			if node != nil {
				node.value += value
			}
		case xml.StartElement:
			if node == nil {
				node = newNode()
				err := node.unmarshal(tt)
				if err != nil {
					return nil, err
				}
			} else {
				childNode, childErr := unmarshalNode(xmlParser, tok)
				if childErr != nil {
					return nil, childErr
				}

				if childNode != nil {
					node.add(childNode)
				} else {
					return
				}
			}
		case xml.EndElement:
			return
		}
	}

	return
}

func getNextToken(xmlParser *xml.Parser) (tok xml.Token, err os.Error) {
	if tok, err = xmlParser.Token(); err != nil {
		if err == os.EOF {
			err = nil
			return
		}
		return
	}

	return
}
