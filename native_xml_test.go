package native_xml_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-xml/native_xml"
)

var xmlstr string = `
<?xml version="1.0" encoding="gb2312"?>
<!DOCTYPE Test Xml "Test.dtd">
<Root root1="rt1" root2="rt2">
  <!-- Comment Normal XML Node -->
  <NodeNormal>ValueNormal</NodeNormal>
  <!-- CDATA -->
  <NodeCDATA><![CDATA[CDATA3001]]></NodeCDATA>
  <!-- attribute -->
  <NodeAttribute ID="Attr_ID" Name="Attr_Name">ValueAttribute</NodeAttribute>
  <Items>
    <!-- Nodes -->
	<Item1>ValueItem1</Item1>
	<Item2>ValueItem2</Item2>
	<Item3>
	  <Item3_1>ValueItem3_1</Item3_1>
	  <Item3_2 ID="Item3_2_ID" Name="Item3_2_Name">
	    <Item3_2_1>ValueItem3_2_1</Item3_2_1>
	  </Item3_2>
	</Item3>
	<Item4>ValueItem4</Item4>
  </Items>
</Root>
`
var editxmlstr string = `
<Root>
  <Head>
  </Head>
  <Body>
  </Body>
</Root>
`
var xmlreadtest = `
<Pub>
	<Item1/>
	<Item2/>
<InstructionCode name="aaa"/>
<Date dname="ASD">aaabbb</Date>
<Time/>
<TradeSource/>
	<Item3/>
</Pub>
`

func shownode(inode *native_xml.TXmlNode) {
	fmt.Println(inode.Name)
	for _, jnode := range inode.Nodes {
		shownode(jnode)
	}
}
func Test_readxml(t *testing.T) {
	nxml := native_xml.NewNativeXml()
	nxml.ReadFromString(xmlreadtest)
	for _, node := range nxml.RootNodes {
		fmt.Println(node.Name)
		for _, inode := range node.Nodes {
			shownode(inode)
		}
	}
	fmt.Println(nxml.WriteToString())
}
func Test_Read_nativexml(t *testing.T) {
	nxml := native_xml.NewNativeXml()
	nxml.ReadFromString(xmlstr)
	nxml.SetXmlFormat(true)
	fmt.Printf("native_xml.WritetoString:\n%s\n", nxml.WriteToString())
	fmt.Printf("XmlNodePath \n%v\n", nxml.XmlNodePath())
	fmt.Printf("XmlNodePathForNode \n%v\n", nxml.XmlNodePathForNode("/Root/Items"))
	if nxml.GetAttribute("/Root", "root1") != "rt1" {
		t.Fatalf("node Root GetAttribute root1 " + nxml.GetAttribute("/Root", "root1") + "!=rt1")
	}
	if nxml.GetAttribute("/Root", "root2") != "rt2" {
		t.Fatalf("node Root GetAttribute root2 " + nxml.GetAttribute("/Root", "root2") + "!=rt2")
	}
	if nxml.GetNodeValueForPath("/Root/NodeNormal") != "ValueNormal" {
		t.Fatalf("node /Root/NodeNormal Value " + nxml.GetNodeValueForPath("/Root/NodeNormal") + "!=ValueNormal")
	}
	if nxml.GetNodeValueForPath("/Root/NodeAttribute") != "ValueAttribute" {
		t.Fatalf("node /Root/NodeAttribute Value " + nxml.GetNodeValueForPath("/Root/NodeAttribute") + "!=ValueAttribute")
	}
	if nxml.GetAttribute("/Root/NodeAttribute", "ID") != "Attr_ID" {
		t.Fatalf("node /Root/NodeAttribute GetAttribute ID " + nxml.GetAttribute("/Root/NodeAttribute", "ID") + "!=Attr_ID")
	}
	if nxml.GetAttribute("/Root/NodeAttribute", "Name") != "Attr_Name" {
		t.Fatalf("node /Root/NodeAttribute GetAttribute Name " + nxml.GetAttribute("/Root/NodeAttribute", "Name") + "!=Attr_Name")
	}
	if nxml.GetNodeValueForPath("/Root/Items/Item1") != "ValueItem1" {
		t.Fatalf("node /Root/Items/Item1 Value " + nxml.GetNodeValueForPath("/Root/Items/Item1") + "!=ValueItem1")
	}
	if nxml.GetNodeValueForPath("/Root/Items/Item2") != "ValueItem2" {
		t.Fatalf("node /Root/Items/Item2 Value " + nxml.GetNodeValueForPath("/Root/Items/Item2") + "!=ValueItem2")
	}
	if nxml.GetNodeValueForPath("/Root/Items/Item3/Item3_1") != "ValueItem3_1" {
		t.Fatalf("node /Root/Items/Item3/Item3_1 Value " + nxml.GetNodeValueForPath("/Root/Items/Item3/Item3_1") + "!=ValueItem3_1")
	}
	if nxml.GetNodeValueForPath("/Root/Items/Item3/Item3_2/Item3_2_1") != "ValueItem3_2_1" {
		t.Fatalf("node /Root/Items/Item3/Item3_2/Item3_2_1 Value " + nxml.GetNodeValueForPath("/Root/Items/Item3/Item3_2/Item3_2_1") + "!=ValueItem3_2_1")
	}
	if nxml.GetAttribute("/Root/Items/Item3/Item3_2", "ID") != "Item3_2_ID" {
		t.Fatalf("node /Root/Items/Item3/Item3_2 GetAttribute ID " + nxml.GetAttribute("/Root/Items/Item3/Item3_2", "ID") + "!=Item3_2_ID")
	}
	if nxml.GetAttribute("/Root/Items/Item3/Item3_2", "Name") != "Item3_2_Name" {
		t.Fatalf("node /Root/Items/Item3/Item3_2 GetAttribute Name " + nxml.GetAttribute("/Root/Items/Item3/Item3_2", "Name") + "!=Item3_2_Name")
	}
	if nxml.GetNodeValueForPath("/Root/Items/Item4") != "ValueItem4" {
		t.Fatalf("node /Root/Items/Item4 Value " + nxml.GetNodeValueForPath("/Root/Items/Item4") + "!=ValueItem4")
	}
}
func Test_Edit_nativexml(t *testing.T) {
	nxml := native_xml.NewNativeXml()
	nxml.ReadFromString(editxmlstr)
	if !nxml.AddNodeForPath("/Root/Items/Item5") {
		t.Fatalf("AddNodeForPath node /Root/Items/Item5")
	}
	if !nxml.SetNodeValueForPath("/Root/Items/Item5", "ValueItem5") {
		t.Fatalf("SetNodeValueForPath node /Root/Items/Item5")
	}
	if nxml.GetNodeValueForPath("/Root/Items/Item5") != "ValueItem5" {
		t.Fatalf("GetNodeValueForPath node /Root/Items/Item5 value " + nxml.GetNodeValueForPath("/Root/Items/Item5") + "!=ValueItem5")
	}
	if !nxml.SetNodeValueForPath("/Root/Head", "HeadValue") {
		t.Fatalf("SetNodeValueForPath node /Root/Head")
	}
	if nxml.GetNodeValueForPath("/Root/Head") != "HeadValue" {
		t.Fatalf("GetNodeValueForPath node /Root/Head value " + nxml.GetNodeValueForPath("/Root/Head") + "!=HeadValue")
	}
	if !nxml.SetAttribute("/Root/Head", "Ver", "1.0") {
		t.Fatalf("SetAttribute node /Root/Head")
	}
	if nxml.GetAttribute("/Root/Head", "Ver") != "1.0" {
		t.Fatalf("GetAttribute node /Root/Head Ver " + nxml.GetAttribute("/Root/Head", "Ver") + "!=1.0")
	}
	if nxml.XMLNodeForPath("/Root/Head") == nil {
		t.Fatalf("XMLNodeForPath node /Root/Head")
	}
	if !nxml.AddNodeForPathS("/Root/Head", "PathS") {
		t.Fatalf("AddNodeForPath node /Root/Head/PathS")
	}
	var tmpstr string = `<recode>
	<item1>valueitem1</item1>
	</recode>
	`
	if !nxml.AddNodeForPathB("/Root/Body", bytes.NewBuffer([]byte(tmpstr))) {
		t.Fatalf("AddNodeForPathB node /Root/Body/recode/item1")
	}
	tmpNode := native_xml.TXmlNode{Attributes: make(map[string]string),
		Nodes:  make(map[int]*native_xml.TXmlNode),
		Name:   "recodeN",
		NodeID: 0,
		Value:  "ValuerecodeN"}
	if !nxml.AddNodeForPathN("/Root/Body", tmpNode) {
		t.Fatalf("AddNodeForPathN node /Root/Body/recodeN")
	}
	tmpRepNode := native_xml.TXmlNode{Attributes: make(map[string]string),
		Nodes:  make(map[int]*native_xml.TXmlNode),
		Name:   "tmpRepNode",
		NodeID: 0,
		Value:  "ValuetmpRepNode"}
	if !nxml.ReplaceNode("/Root/Items/Item5", &tmpRepNode) {
		t.Fatalf("ReplaceNode node /Root/Items/Item5 with tmpRepNode")
	}
	fmt.Println("/Root/Items/tmpRepNode TreeDepth=" + strconv.Itoa(nxml.XMLNodeForPath("/Root/Items/tmpRepNode").TreeDepth()))
	fmt.Println("xfCompact:\n" + nxml.WriteToString())
	nxml.SetXmlFormat(true)
	fmt.Println("xfReadable:\n" + nxml.WriteToString())
}
