// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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

func (this *xmlNode) unmarshal(startEl xml.StartElement) error {
	this.name = startEl.Name.Local

	for _, v := range startEl.Attr {
		_, alreadyExists := this.attributes[v.Name.Local]
		if alreadyExists {
			return errors.New("Tag '" + this.name + "' has duplicated attribute: '" + v.Name.Local + "'")
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

func (this *xmlNode) hasChildren() bool {
	return this.children != nil && len(this.children) > 0
}

//=============================================

func unmarshalConfig(reader io.Reader) (*xmlNode, error) {
	xmlParser := xml.NewDecoder(reader)

	config, err := unmarshalNode(xmlParser, nil)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("Xml has no content")
	}

	nextConfigEntry, err := unmarshalNode(xmlParser, nil)
	if nextConfigEntry != nil {
		return nil, errors.New("Xml contains more than one root element")
	}

	return config, nil
}

func unmarshalNode(xmlParser *xml.Decoder, curToken xml.Token) (node *xmlNode, err error) {
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
			err = errors.New(tt.Error())
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

func getNextToken(xmlParser *xml.Decoder) (tok xml.Token, err error) {
	if tok, err = xmlParser.Token(); err != nil {
		if err == io.EOF {
			err = nil
			return
		}
		return
	}

	return
}
