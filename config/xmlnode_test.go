// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"testing"
	"strings"
	//"fmt"
	"reflect"
)

var testEnv *testing.T

/*func TestWrapper(t *testing.T) {
	testEnv = t

	s := "<a d='a'><g m='a'></g><g h='t' j='kk'></g></a>"
	reader := strings.NewReader(s)
	config, err := unmarshalConfig(reader)
	if err != nil {
		testEnv.Error(err)
		return
	}

	printXml(config, 0)
}

func printXml(node *xmlNode, level int) {
	indent := strings.Repeat("\t", level)
	fmt.Print(indent + node.name)
	for key, value := range node.attributes {
		fmt.Print(" " + key + "/" + value)
	}
	fmt.Println()

	for _, child := range node.children {
		printXml(child, level+1)
	}
}*/

var xmlNodeTests []xmlNodeTest

type xmlNodeTest struct {
	testName      string
	inputXml      string
	expected      interface{}
	errorExpected bool
}

func getXmlTests() []xmlNodeTest {
	if xmlNodeTests == nil {
		xmlNodeTests = make([]xmlNodeTest, 0)

		testName := "Simple test"
		testXml := `<a></a>`
		testExpected := newNode()
		testExpected.name = "a"
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Multiline test"
		testXml =
			`
<a>
</a>
`
		testExpected = newNode()
		testExpected.name = "a"
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Multiline test #2"
		testXml =
			`


<a>

</a>

`
		testExpected = newNode()
		testExpected.name = "a"
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Incorrect names"
		testXml = `< a     ><      /a >`
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, nil, true})

		testName = "Comments"
		testXml =
			`<!-- <abcdef/> -->
<a> <!-- <!--12345-->
</a>
`
		testExpected = newNode()
		testExpected.name = "a"
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Multiple roots"
		testXml = `<a></a><b></b>`
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, nil, true})

		testName = "Multiple roots + incorrect xml"
		testXml = `<a></a><b>`
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, nil, true})

		testName = "Some unicode and data"
		testXml = `<俄语>данные</俄语>`
		testExpected = newNode()
		testExpected.name = "俄语"
		testExpected.value = "данные"
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Values and children"
		testXml = `<俄语>данные<and_a_child></and_a_child></俄语>`
		testExpected = newNode()
		testExpected.name = "俄语"
		testExpected.value = "данные"
		child := newNode()
		child.name = "and_a_child"
		testExpected.children = append(testExpected.children, child)
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Just children"
		testXml = `<俄语><and_a_child></and_a_child></俄语>`
		testExpected = newNode()
		testExpected.name = "俄语"
		child = newNode()
		child.name = "and_a_child"
		testExpected.children = append(testExpected.children, child)
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})

		testName = "Mixed test"
		testXml = `<俄语 a="1" b="2.13" c="abc"><child abc="bca"/><child abc="def"></child></俄语>`
		testExpected = newNode()
		testExpected.name = "俄语"
		testExpected.attributes["a"] = "1"
		testExpected.attributes["b"] = "2.13"
		testExpected.attributes["c"] = "abc"
		child = newNode()
		child.name = "child"
		child.attributes["abc"] = "bca"
		testExpected.children = append(testExpected.children, child)
		child = newNode()
		child.name = "child"
		child.attributes["abc"] = "def"
		testExpected.children = append(testExpected.children, child)
		xmlNodeTests = append(xmlNodeTests, xmlNodeTest{testName, testXml, testExpected, false})
	}

	return xmlNodeTests
}

func TestXmlNode(t *testing.T) {

	for _, test := range getXmlTests() {

		reader := strings.NewReader(test.inputXml)
		parsedXml, err := unmarshalConfig(reader)

		if (err != nil) != test.errorExpected {
			t.Errorf("\n%s:\nXml input: %s\nExpected error:%t. Got error: %t\n", test.testName,
				test.inputXml, test.errorExpected, (err != nil))
			if err != nil {
				t.Logf("%s\n", err.Error())
			}
			continue
		}

		if err == nil && !reflect.DeepEqual(parsedXml, test.expected) {
			t.Errorf("\n%s:\nXml input: %s\nExpected: %s. \nGot: %s\n", test.testName, test.inputXml, test.expected, parsedXml)
		}
	}
}
