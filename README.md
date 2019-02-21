# go-xml
native_xml: support with nodepath get or set node value and have more oprater xml method.<br/>
the code from delphi native componet transfer.

example:

package main

import (
	"fmt"
	"github.com/jjjgit/go-xml"
)

func main(){<br/>
	xmlstr:=`<?xml version="1.0" encoding="UTF-8"?><root><row1>rowdata</row1><data/></root>`<br/>
	fmt.Println(xmlstr)<br/>
	xml:=native_xml.NewNativeXml()<br/>
	xml.ReadFromString(xmlstr)<br/>
	fmt.Println(xml.GetNodeValueForPath("/root/row1"))<br/>
	xml.RemoveNode("/root/row1")<br/>
	xml.SetNodeValueForPath("/root/data","setdata")<br/>
	fmt.Println(xml.GetNodeValueForPath("/root/data"))<br/>
	xml.AddNodeForPath("/root/addrow")<br/>
	xml.SetNodeValueForPath("/root/addrow","adddata")<br/>
	fmt.Println(xml.GetNodeValueForPath("/root/addrow"))<br/>
	xmlstr=xml.WriteToString()<br/>
	fmt.Println(xmlstr)<br/>
}<br/>
