package native_xml

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type TXmlElementType int
type TxmlFormatType int

//internal type
type TTagType struct {
	FStart string
	FClose string
	FStyle TXmlElementType
}

const (
	xeNormal      TXmlElementType = iota // Normal element <name {attr}>[value][sub-elements]</name>
	xeComment                            // Comment <!--{comment}-->
	xeCData                              // literal data <![CDATA[{data}]]>
	xeDeclaration                        // XML declareration <?xml{declaration}?>
	xeStyleSheet                         // StyleSheet <?xml-stylesheet{stylesheet}?>
	xeDocType                            // DOCTYPE DTD declaration <!DOCTYPE{spec}>
	xeElement                            // <!ELEMENT> 6
	xeAttList                            // <!ATTLIST>
	xeEntity                             // <!ENTITY>
	xeNotation                           // <!NOTATION>
	xeExclam                             // Any <!data>
	xeQuestion                           // Any <?data?> 11
	xeCharData                           // Character data in a node
	xeUnknown                            // Any <data>
	cTagCount     int             = 12

	xfReadable TxmlFormatType = iota
	xfCompact

	sxeErrorCalcStreamLenght       = "Error while calculating stream length"
	sxeMissingDataInBinaryStream   = "Missing data in binary stream"
	sxeMissingElementName          = "Missing element name"
	sxeMissingCloseTag             = "Missing close tag in element %s"
	sxeMissingDataAfterGreaterThan = "Missing data after \"<\" in element %s"
	sxeMissingLessThanInCloseTag   = "Missing \">\" in close tag of element %s"
	sxeIncorrectCloseTag           = "Incorrect close tag in element %s"
	sxeIllegalCharInNodeName       = "Illegal character in node name \"%s\""
	sxeMoreThanOneRootElement      = "More than one root element found in xml"
	sxeMoreThanOneDeclaration      = "More than one xml declaration found in xml"
	sxeDeclarationMustBeFirstElem  = "Xml declaration must be first element"
	sxeMoreThanOneDoctype          = "More than on doctype declaration found in root"
	sxeDoctypeAfterRootElement     = "Doctype declaration found after root element"
	sxeNoRootElement               = "No root element found in xml"
	sxeIllegalElementType          = "Illegal element type"
	sxeCDATAInRoot                 = "No CDATA allowed in root"
	sxeRootElementNotDefined       = "XML root element not defined"
	sxeCodecStreamNotAssigned      = "Encoding stream unassigned"
	sxeUnsupportedEncoding         = "Unsupported string encoding"
	sxeCannotReadCodecForWriting   = "Cannot read from a conversion stream opened for writing"
	sxeCannotWriteCodecForReading  = "Cannot write to an UTF stream opened for reading"
	sxeCannotReadMultipleChar      = "Cannot read multiple chars from conversion stream at once"
	sxeCannotPerformSeek           = "Cannot perform seek on codec stream"
	sxeCannotSeekBeforReadWrite    = "Cannot seek before reading or writing in conversion stream"
	sxeCannotSeek                  = "Cannot perform seek in conversion stream"
	sxeCannotWriteToOutputStream   = "Cannot write to output stream"
	sxeXmlNodeNotAssigned          = "XML Node is not assigned"
	sxeCannotConvertToBool         = "Cannot convert value to bool"
	sxeCannotCovertToFloat         = "Cannot convert value to float"
	sxeSignificantDigitsOutOfRange = "Significant digits out fo range"
)

var (
	cQuoteChars   = "\"'"              //[2]byte{'"','\''}
	cControlChars = "\x09\x0A\x0D\x20" //{Tab,Lf,CR,Space}

	cTags = [cTagCount]TTagType{
		//the order is important here;the items are searched for in appearing order
		{FStart: "<![CDATA[", FClose: "]]>", FStyle: xeCData},
		{FStart: "<!DOCTYPE", FClose: ">", FStyle: xeDocType},
		{FStart: "<!ELEMENT", FClose: ">", FStyle: xeElement}, //2
		{FStart: "<!ATTLIST", FClose: ">", FStyle: xeAttList}, //3
		{FStart: "<!ENTITY", FClose: ">", FStyle: xeEntity},
		{FStart: "<!NOTATION", FClose: ">", FStyle: xeNotation},
		{FStart: "<?xml-stylesheet", FClose: "?>", FStyle: xeStyleSheet},
		{FStart: "<?xml", FClose: "?>", FStyle: xeDeclaration}, //7
		{FStart: "<!--", FClose: "-->", FStyle: xeComment},
		{FStart: "<!", FClose: ">", FStyle: xeExclam},
		{FStart: "<?", FClose: "?>", FStyle: xeQuestion},
		{FStart: "<", FClose: ">", FStyle: xeNormal}, //11
		// direct tags are derived from Normal tags by checking for the />
	}
)

type TsdSurplusReader struct {
	Reader  *bytes.Reader
	Surplus string
}

func (this *TsdSurplusReader) ReadChar() (Ch byte, readlen int) {
	if len(this.Surplus) > 0 {
		Ch = this.Surplus[0]
		this.Surplus = string(this.Surplus[1:])
		readlen = 1
	} else {
		var err error
		Ch, err = this.Reader.ReadByte()
		if err != nil {
			readlen = 0
		} else {
			readlen = 1
		}
	}
	return Ch, readlen
}
func (this *TsdSurplusReader) ReadCharSkipBlanks() (Ch byte, b bool) {
	for exec := true; exec; {
		//Read character,exit if none available
		i := 0
		if Ch, i = this.ReadChar(); i == 0 {
			//Skip if in controlchars
			return Ch, false
		}
		exec = func() bool {
			if strings.Index(cControlChars, string(Ch)) >= 0 {
				return false
			}
			return false
		}()
	}
	return Ch, true
}

// Xml Node
type TXmlNode struct {
	Attributes  map[string]string //List with attributes
	document    *TNativeXml       //*Only Root node need set .Pointer to parent Xml Document
	ElementType TXmlElementType   //The type of element
	Name        string            //The element name
	Nodes       map[int]*TXmlNode //These are the child elements
	Parent      *TXmlNode         //Pointer to parent element
	Tag         int               //A value the developer can use
	Value       string            // The *escaped* value
	MaxNodeID   int               // Node item id count
	NodeID      int               // Node at level globle id
}

func NewXmlNode(nodename string) *TXmlNode {
	return &TXmlNode{Attributes: make(map[string]string),
		Nodes:  make(map[int]*TXmlNode),
		Name:   nodename,
		NodeID: 0,
		Value:  ""}
}
func (this *TXmlNode) Document() *TNativeXml {
	if this.Parent != nil {
		return this.Parent.Document()
	} else {
		return this.document
	}
	return nil
}
func (this *TXmlNode) TreeDepth() int {
	//The node level
	if this.Parent != nil {
		return this.Parent.TreeDepth() + 1
	} else {
		return 0
	}
}
func (this *TXmlNode) NodeCount() int {
	return len(this.Nodes)
}
func (this *TXmlNode) ParseTag(AValue string, TagStart, TagClose int) {
	//Create a list to hold string items
	ParseAttributes(AValue, TagStart, TagClose, this.Attributes)
	//Determine name,attributes or value for each element type
	switch this.ElementType {
	case xeDeclaration:
		this.Name = "xml"
	case xeStyleSheet:
		this.Name = "xml-stylesheet"
		//We also set this as the value for use in "StyleSheetString"
		this.Value = AValue[TagStart : TagClose-TagStart]
	}
}
func (this *TXmlNode) NodeAdd(ANode *TXmlNode) int {
	if ANode != nil {
		ANode.Parent = this
		if this.Nodes == nil {
			this.Nodes = make(map[int]*TXmlNode)
		}
		this.MaxNodeID++
		ANode.NodeID = this.MaxNodeID
		this.Nodes[this.MaxNodeID] = ANode
		return this.MaxNodeID
	} else {
		return -1
	}
}
func (this *TXmlNode) AddCharDataNode(ANodeValue string) {
	//Add all text up till now as xeCharData
	ANodeValue = strings.Trim(ANodeValue, " ")
	ANodeValue = strings.Trim(ANodeValue, "\x0A")
	ANodeValue = strings.Trim(ANodeValue, "\x0D")
	this.Value = ANodeValue
}
func (this *TXmlNode) ReadFromString(AValue string) {
	rd := &bytes.Reader{}
	rd.Read([]byte(AValue))
	this.ReadFromStream(rd)
}
func (this *TXmlNode) ReadFromStream(S *bytes.Reader) {
	//Read the node from the starting "<" until the closing ">" from the stream in S.
	ANodeValue := new(bytes.Buffer)
	HasCR := false
	HasSubTags := false
	var (
		err    error
		Ch     byte
		bret   bool
		AValue string
	)
	Reader := &TsdSurplusReader{Reader: S}
	//Trailing blanks/controls chars?
	if Ch, bret = Reader.ReadCharSkipBlanks(); !bret {
		return
	}
	//What is it? Tag is End or Start
	if Ch == '<' {
		// A tag - which one?
		ATagIndex := ReadOpenTag(Reader)
		if ATagIndex >= 0 {
			this.ElementType = cTags[ATagIndex].FStyle
			switch this.ElementType {
			case xeNormal, xeDeclaration, xeStyleSheet:
				//These tags we will process
				AValue, _ = ReadStringFromStreamUntil(Reader, cTags[ATagIndex].FClose, true)
				ALength := len(AValue)
				IsDirect := false
				if this.ElementType == xeNormal {
					if (ALength > 0) && (AValue[ALength-1] == '/') {
						//Is it a direct tag?
						ALength--
						IsDirect = true
						AValue = AValue[:ALength]
					}
					for i := 0; i < ALength; i++ {
						if strings.Index(cControlChars, string(AValue[i])) >= 0 {
							this.Name = strings.Trim(AValue[:i], " ")
							AValue = AValue[i:]
							break
						} else {
							this.Name = strings.Trim(AValue, " ")
						}
					}
					if this.Name == "" {
						this.Name = strings.Trim(AValue, " ")
					}
				}
				ALength = len(AValue)

				this.ParseTag(AValue, 0, ALength-1)
				//Now the tag can be a direct close - in that case we're finished
				if IsDirect || this.ElementType == xeDeclaration || this.ElementType == xeStyleSheet {
					return
				}
				//Process reset of tag
				for {
					//Read character from stream
					if Ch, err = S.ReadByte(); err != nil {
						panic(errors.New(fmt.Sprintf(sxeMissingCloseTag, this.Name)))
					}
					//Is there a subtag?
					if Ch == '<' {
						if Ch, bret = Reader.ReadCharSkipBlanks(); !bret {
							panic(errors.New(fmt.Sprintf(sxeMissingDataAfterGreaterThan, this.Name)))
						}
						if Ch == '/' {
							//This seems our closing tag
							if AValue, bret = ReadStringFromStreamUntil(Reader, ">", true); !bret {
								panic(errors.New(fmt.Sprintf(sxeMissingLessThanInCloseTag, this.Name)))
							}
							if strings.Compare(strings.Trim(AValue, " "), this.Name) != 0 {
								panic(errors.New(fmt.Sprintf(sxeIncorrectCloseTag, this.Name)))
							}
							AValue = ""
							break
						} else {
							//Add all text up till now as xeCharData
							this.AddCharDataNode(ANodeValue.String())
							ANodeValue.Reset()
							//Reset this HasCR flag if we add node ,we only want to detect
							//The CR after last subnode
							HasCR = false
							//This is a subtag... so create it and let it process
							HasSubTags = true
							S.Seek(-2, io.SeekCurrent)
							ANode := &TXmlNode{Attributes: make(map[string]string),
								Nodes: make(map[int]*TXmlNode)}
							this.NodeAdd(ANode)
							ANode.ReadFromStream(S)
						}
					} else {
						//If we detect a CR we will set the flag.This will signal the fact
						//That this XML file was saved with xfReadable
						if Ch == 0x0D {
							HasCR = true
						}
						//Add the character to the node value buffer.
						ANodeValue.WriteByte(Ch)
					}
				}
				//Add all text up till now as xeText
				this.AddCharDataNode(ANodeValue.String())
				ANodeValue.Reset()
				//Check CharData nodes,remove trailing CRLF + indentation if we
				//were in xfReadable mode
				if HasSubTags && HasCR {
					for _, v := range this.Nodes {
						if v.ElementType == xeCharData {
							AClose := strings.IndexAny(v.Value, "\x0D\x0A")
							if AClose < 0 {
								AClose = len(v.Value)
							}
							v.Value = v.Value[:AClose]
						}
					}
				}
				// If this first node is xeCharData we use it as ValueDirect
			case xeDocType:
				this.Name = "DTD"
				AValue, _ = ReadStringFromStreamUntil(Reader, cTags[ATagIndex].FClose, false)
				this.Value = AValue
			//Parse DTD
			case xeElement, xeAttList, xeEntity, xeNotation:
				//DTD elements
				AValue, _ = ReadStringFromStreamWithQuotes(Reader, cTags[ATagIndex].FClose)
				ALength := len(AValue)
				Words := make(map[string]string)
				ParseAttributes(AValue, 0, ALength-1, Words)
				for k, v := range Words {
					if len(this.Name) == 0 {
						this.Name = k
					}
					this.Value += k + "=" + v + "\x0D\x0A"
				}
			default:
				switch this.ElementType {
				case xeComment:
					this.Name = "Comment"
				case xeCData:
					this.Name = "CData"
				case xeExclam:
					this.Name = "Special"
				case xeQuestion:
					this.Name = "Special"
				default:
					this.Name = "Unknown"
				}
				//In these cases just get all data up till the closing tag
				AValue, _ = ReadStringFromStreamUntil(Reader, cTags[ATagIndex].FClose, false)
				this.Value = AValue
			} //case
		}
	}
}
func (this *TXmlNode) GetIndent() string {
	if this.Document != nil {
		switch this.Document().XmlFormat {
		case xfCompact:
			return ""
		case xfReadable:
			return strings.Repeat(this.Document().IndentString, this.TreeDepth())
		}
	}
	return ""
}
func (this *TXmlNode) GetLineFeed() string {
	if this.Document != nil {
		switch this.Document().XmlFormat {
		case xfCompact:
			return ""
		case xfReadable:
			return "\x0D\x0A"
		default:
			return "\x0A"
		}
	}
	return ""
}
func (this *TXmlNode) UseFullNodes() bool {
	if this.Document() != nil {
		return this.Document().UseFullNodes
	}
	return false
}
func (this *TXmlNode) QualifyAsDirectNode() bool {
	//If this node qualifies as a direct node when writing ,we return true.
	//A direct node may have attributes,but no value or subnodes.Furhtermore
	//The root node will never be displayed as a direct node.
	return (len(this.Value) == 0) &&
		(this.NodeCount() == 0) &&
		(this.ElementType == xeNormal) &&
		!this.UseFullNodes() &&
		(this.TreeDepth() > 0)
}
func (this *TXmlNode) DeclarationWriteInnerTag() string {
	//Write the inner part of the tag,the one that contains the attributes
	//Attributes
	val := ""
	//Do not write empty attributes
	for k, v := range this.Attributes {
		if strings.ToLower(k) == "version" {
			val = " " + k + "=\"" + v + "\" " + val
		} else {
			val += " " + k + "=\"" + v + "\""
		}
	}
	//End of tag - direct nodes get an extra "/"
	if this.QualifyAsDirectNode() {
		val += "/"
	}
	return val
}
func (this *TXmlNode) WriteInnerTag() string {
	//Write the inner part of the tag,the one that contains the attributes
	//Attributes
	val := ""
	//Do not write empty attributes
	for k, v := range this.Attributes {
		val += " " + k + "=\"" + v + "\""
	}
	//End of tag - direct nodes get an extra "/"
	if this.QualifyAsDirectNode() {
		val += "/"
	}
	return val
}
func (this *TXmlNode) WriteToString() string {
	buf := &bytes.Buffer{}
	this.WriteToStream(buf)
	return buf.String()
}

func (this *TXmlNode) WriteToStream(S *bytes.Buffer) {
	AIndent := this.GetIndent()
	ALineFeed := this.GetLineFeed()
	NodeCount := this.NodeCount()
	//Write indent
	ALine := AIndent
	//Write the node - disinguish node type
	switch this.ElementType {
	case xeDeclaration: //Xml declaration <?xml{declaration}?>
		ALine = AIndent + fmt.Sprintf("<?xml%s?>", this.DeclarationWriteInnerTag())
	case xeStyleSheet: //StyeSheet <?xml-stylesheet{stylesheet}?>
		ALine = AIndent + fmt.Sprintf("<?xml-stylesheet%s?>", this.WriteInnerTag())
	case xeDocType:
		if NodeCount == 0 {
			ALine = AIndent + fmt.Sprintf("<!DOCTYPE %s>", this.Value)
		} else {
			ALine = AIndent + fmt.Sprintf("<!DOCTYPE %s[", this.Value) + ALineFeed
			WriteStringToStream(S, ALine)
			for _, v := range this.Nodes {
				v.WriteToStream(S)
				WriteStringToStream(S, ALineFeed)
			}
			ALine = "]>"
		}
	case xeElement:
		ALine = AIndent + fmt.Sprintf("<!ELEMENT %s %s>", this.Name, this.Value)
	case xeAttList:
		ALine = AIndent + fmt.Sprintf("<!ATTLIST %s %s>", this.Name, this.Value)
	case xeEntity:
		ALine = AIndent + fmt.Sprintf("<!ENTITY %s %s>", this.Name, this.Value)
	case xeNotation:
		ALine = AIndent + fmt.Sprintf("<!NOTATION %s %s>", this.Name, this.Value)
	case xeComment:
		ALine = AIndent + fmt.Sprintf("<!--%s-->", this.Value)
	case xeCData:
		ALine = AIndent + fmt.Sprintf("<![CDATA[%s]]>", this.Value)
	case xeExclam:
		ALine = AIndent + fmt.Sprintf("<!%s>", this.Value)
	case xeQuestion:
		ALine = AIndent + fmt.Sprintf("<?%s?>", this.Value)
	case xeCharData:
		ALine = this.Value
	case xeUnknown:
		ALine = AIndent + fmt.Sprintf("<%s>", this.Value)
	case xeNormal:
		//Write tag
		ALine = AIndent + fmt.Sprintf("<%s%s>", this.Name, this.WriteInnerTag())
		//Write value (if Any)
		ALine += this.Value
		if NodeCount > 0 {
			//..and a linefeed
			ALine += ALineFeed
		}
		WriteStringToStream(S, ALine)
		//Write child element
		for _, v := range this.Nodes {
			v.WriteToStream(S)
			if v.ElementType != xeCharData {
				WriteStringToStream(S, ALineFeed)
			}
		}
		//Write end tag
		ALine = ""
		if !this.QualifyAsDirectNode() {
			if NodeCount > 0 {
				ALine = AIndent
			}
			ALine += fmt.Sprintf("</%s>", this.Name)
		}
	default:
		panic(errors.New(sxeIllegalElementType))
	}
	WriteStringToStream(S, ALine)
}
func (this *TXmlNode) HasAttribute(AName string) bool {
	_, b := this.Attributes[AName]
	return b
}
func (this *TXmlNode) IsEmpty() bool {
	return (len(this.Value) == 0) && (this.NodeCount() == 0) && len(this.Attributes) == 0
}
func (this *TXmlNode) IsClear() bool {
	return len(this.Name) == 0 && this.IsEmpty()
}

func (this *TXmlNode)NameToValue(name string) string{
	for _,v:=range this.Nodes{
		if v.Name==name{
			return v.Value
		}
	}
	return ""
}

//Xml Operation
type TNativeXml struct {
	XmlString      string
	XmlFormat      TxmlFormatType
	IndentString   string
	UseFullNodes   bool
	XmlRoot        *TXmlNode
	RootNodes      map[TXmlElementType]*TXmlNode
	ParserWarnings bool
}

func (this *TNativeXml) SetXmlFormat(xftype bool) {
	if xftype {
		this.XmlFormat = xfReadable
	} else {
		this.XmlFormat = xfCompact
	}
}
func (this *TNativeXml) LineFeed() string {
	switch this.XmlFormat {
	case xfCompact:
		return ""
	case xfReadable:
		return "\x0D\x0A"
	default:
		return "\x0A"
	}
}
func (this *TNativeXml) WriteToStream(S *bytes.Buffer) {
	if this.RootNodes == nil && this.ParserWarnings {
		panic(errors.New(sxeRootElementNotDefined))
	}
	//Write the Xml declaration <?xml{declaration}?>
	for k, v := range this.RootNodes {
		if k == xeDeclaration {
			v.WriteToStream(S)
			WriteStringToStream(S, this.LineFeed())
		}
	}
	//Write to XML DOCTYPE DTD declaration <!DOCTYPE{spec}>
	for k, v := range this.RootNodes {
		if k == xeDocType {
			v.WriteToStream(S)
			WriteStringToStream(S, this.LineFeed())
		}
	}
	//Write the root node
	for k, v := range this.RootNodes {
		if k == xeNormal || k == xeCData {
			v.WriteToStream(S)
			WriteStringToStream(S, this.LineFeed())
		}
	}
}
func (this *TNativeXml) WriteToString() string {
	buf := new(bytes.Buffer)
	this.WriteToStream(buf)
	return buf.String()
}
func (this *TNativeXml) LoadFromFile(FileName string) {
	f, err := os.Open(FileName)
	if err != nil {
		return
	}
	bufarr := [4096]byte{}
	var fbuf []byte = bufarr[:]
	buf := new(bytes.Buffer)
	for i, err := f.Read(fbuf); true; i, err = f.Read(fbuf) {
		if err == io.EOF {
			if i > 0 {
				buf.Write(fbuf[:])
			}
			break
		} else if err != nil {
			return
		}
		if i > 0 {
			buf.Write(fbuf[:])
		}
	}
	this.ReadFromStream(buf)
}
func (this *TNativeXml) ReadFromString(AValue string) {
	this.ReadFromStream(bytes.NewBuffer([]byte(AValue)))
}
func (this *TNativeXml) ReadFromStream(S *bytes.Buffer) {
	this.XmlString = S.String()
	//Clear the old root nodes - we do not reset the defaults
	this.RootNodes = make(map[TXmlElementType]*TXmlNode)
	Reader := bytes.NewReader(S.Bytes())
	for Reader.Len() > 0 {
		ANode := &TXmlNode{Attributes: make(map[string]string),
			document: this,
			Nodes:    make(map[int]*TXmlNode)}
		ANode.ReadFromStream(Reader)
		//XML declaration
		if ANode.ElementType == xeDeclaration {
			//if has "encoding" node ,check encoding and encode content
		}
		//Skip clear nodes
		if !ANode.IsClear() {
			if ANode.ElementType == xeNormal {
				this.XmlRoot = ANode
			}
			this.RootNodes[ANode.ElementType] = ANode
		}
	}
	//Do some checks
	NormalCount := 0
	DeclarationCount := 0
	DoctypeCount := 0
	CDataCount := 0
	NormalPos := -1
	DoctypePos := -1
	for _, v := range this.RootNodes {
		//Count normal elements - there may be only one
		switch v.ElementType {
		case xeNormal:
			NormalCount++
		case xeDeclaration:
			DeclarationCount++
		case xeDocType:
			DoctypeCount++
		case xeCData:
			CDataCount++
		}
	}
	//We *must* have a root node
	if NormalCount == 0 {
		panic(errors.New(sxeNoRootElement))
	}
	//Do some validation if we allow parser warnings
	if this.ParserWarnings {
		//Check for more than one root node
		if NormalCount > 1 {
			panic(errors.New(sxeMoreThanOneRootElement))
		}
		//Check for more than one xml declaration
		if DeclarationCount > 1 {
			panic(errors.New(sxeMoreThanOneDeclaration))
		}
		//Check for more than one DTD
		if DoctypeCount > 1 {
			panic(errors.New(sxeMoreThanOneDoctype))
		}
		//Check if DTD is after root, this is not allowed
		if (DoctypeCount == 1) && (DoctypePos > NormalPos) {
			panic(errors.New(sxeDoctypeAfterRootElement))
		}
		//No CDATA in root allowed
		if CDataCount > 0 {
			panic(errors.New(sxeCDATAInRoot))
		}
	}
}
func (this *TNativeXml) getpath(v *TXmlNode, parent string, nodepath *[]string) {
	if v.ElementType == xeCData {
		*nodepath = append(*nodepath, parent)
	} else if len(v.Nodes) > 0 {
		for _, v1 := range v.Nodes {
			if v1.ElementType == xeNormal || v1.ElementType == xeCData {
				this.getpath(v1, parent+"/"+v.Name, nodepath)
			}
		}
	} else {
		if v.ElementType == xeNormal || v.ElementType == xeCData {
			*nodepath = append(*nodepath, parent+"/"+v.Name)
		}
	}
}
func (this *TNativeXml) XmlNodePath() []string {
	if this.XmlRoot == nil {
		panic(errors.New(sxeNoRootElement))
	}
	rootcount := 0
	nodepath := make([]string, 0)
	for k, v := range this.RootNodes {
		if k == xeNormal {
			rootcount++
			this.getpath(v, "", &nodepath)
		}
	}
	if rootcount != 1 {
		panic(errors.New(sxeMoreThanOneRootElement))
	}
	return nodepath
}
func (this *TNativeXml) XmlNodePathForNode(NodePath string) []string {
	if this.XmlRoot == nil {
		panic(errors.New(sxeNoRootElement))
	}
	findnode := this.findNodeForPath(NodePath)
	if findnode == nil {
		return make([]string, 0)
	}

	nodepath := make([]string, 0)
	for _, v := range findnode.Nodes {
		this.getpath(v, "/"+findnode.Name, &nodepath)
	}
	return nodepath
}
func (this *TNativeXml) findNodeForName(NodeName string, Node *TXmlNode) *TXmlNode {
	for _, v := range Node.Nodes {
		if (v.ElementType == xeNormal || v.ElementType == xeCData) && v.Name == NodeName {
			return v
		}
	}
	return nil
}
func (this *TNativeXml) findNodeForPath(ParentPath string) *TXmlNode {
	spath := strings.Replace(ParentPath, " ", "", -1)
	path := strings.Split(spath, "/")
	var findnode *TXmlNode
	for _, v := range path {
		if v == "" {
			continue
		}
		if findnode == nil {
			if this.XmlRoot.Name == v {
				findnode = this.XmlRoot
				continue
			} else {
				return nil
			}
		}
		findnode = this.findNodeForName(v, findnode)
		if findnode == nil {
			return nil
		}
	}
	return findnode
}
func (this *TNativeXml) AddNodeForPathN(ParentPath string, Child TXmlNode) bool {
	findnode := this.findNodeForPath(ParentPath)
	if findnode != nil {
		findnode.MaxNodeID++
		Child.NodeID = findnode.MaxNodeID
		Child.Parent = findnode
		findnode.Nodes[Child.NodeID] = &Child
	}
	return findnode != nil
}
func (this *TNativeXml) AddNodeForPathS(ParentPath string, Child string) bool {
	findnode := this.findNodeForPath(ParentPath)
	if findnode != nil {
		findnode.MaxNodeID++
		findnode.Nodes[findnode.MaxNodeID] = &TXmlNode{Attributes: make(map[string]string),
			Nodes:  make(map[int]*TXmlNode),
			Name:   Child,
			NodeID: findnode.MaxNodeID,
			Parent: findnode}
	}
	return findnode != nil
}
func (this *TNativeXml) AddNodeForPathB(ParentPath string, Child *bytes.Buffer) bool {
	findnode := this.findNodeForPath(ParentPath)
	if findnode != nil {
		newnativexml := NewNativeXml()
		newnativexml.ReadFromStream(Child)
		if newnativexml.XmlRoot != nil {
			findnode.MaxNodeID++
			newnativexml.XmlRoot.NodeID = findnode.MaxNodeID
			newnativexml.XmlRoot.Parent = findnode
			findnode.Nodes[findnode.MaxNodeID] = newnativexml.XmlRoot
		} else {
			return false
		}
	}
	return findnode != nil
}
func (this *TNativeXml) AddNodeForPath(Path string) bool {
	spath := strings.Replace(Path, " ", "", -1)
	path := strings.Split(spath, "/")
	var findnode, profindnode *TXmlNode
	for _, v := range path {
		if v == "" {
			continue
		}
		if findnode == nil {
			if this.XmlRoot == nil {
				findnode = &TXmlNode{Attributes: make(map[string]string),
					document: this,
					Nodes:    make(map[int]*TXmlNode),
					Name:     v,
					NodeID:   0}
				this.RootNodes[xeNormal] = findnode
				this.XmlRoot = findnode
				continue
			} else if this.XmlRoot.Name == v {
				findnode = this.XmlRoot
				continue
			} else {
				return false
			}
		}
		profindnode = findnode
		findnode = this.findNodeForName(v, findnode)
		if findnode == nil {
			profindnode.MaxNodeID++
			findnode = &TXmlNode{Attributes: make(map[string]string),
				Nodes:  make(map[int]*TXmlNode),
				Name:   v,
				NodeID: profindnode.MaxNodeID,
				Parent: profindnode}
			profindnode.Nodes[profindnode.MaxNodeID] = findnode
		} else {
			profindnode = findnode
		}
	}
	return findnode != nil
}
func (this *TNativeXml) XMLNodeForPath(FindPath string) *TXmlNode {
	return this.findNodeForPath(FindPath)
}
func (this *TNativeXml) SetNodeValueForPath(FindPath, Value string) bool {
	findnode := this.findNodeForPath(FindPath)
	if findnode == nil {
		return false
	} else {
		findnode.Value = Value
		return true
	}
}
func (this *TNativeXml) GetNodeValueForPath(FindPath string) string {
	findnode := this.findNodeForPath(FindPath)
	if findnode == nil {
		return ""
	} else {
		return findnode.Value
	}
}
func (this *TNativeXml) ReplaceNode(FindPath string, Node *TXmlNode) bool {
	findnode := this.findNodeForPath(FindPath)
	var profindnode *TXmlNode
	if findnode != nil {
		profindnode = findnode.Parent
	} else {
		return false
	}
	if profindnode != nil {
		Node.Parent = profindnode
		profindnode.Nodes[findnode.NodeID] = Node
		return true
	} else {
		return false
	}
}
func (this *TNativeXml) GetAttribute(FindPath, AttrName string) string {
	findnode := this.findNodeForPath(FindPath)
	if findnode == nil {
		return ""
	} else {
		return findnode.Attributes[AttrName]
	}
}
func (this *TNativeXml) SetAttribute(FindPath, AttrName, AttrValue string) bool {
	findnode := this.findNodeForPath(FindPath)
	if findnode != nil {
		findnode.Attributes[AttrName] = AttrValue
		return true
	} else {
		return false
	}
}
func (this *TNativeXml) RemoveNode(FindPath string) bool {
	findnode := this.findNodeForPath(FindPath)
	if findnode != nil {
		key := findnode.NodeID
		findnode = findnode.Parent
		delete(findnode.Nodes, key)
		return true
	} else {
		return false
	}
}
