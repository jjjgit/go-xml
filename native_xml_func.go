package native_xml

import (
	"bytes"
	"strings"
)

func NewNativeXml() *TNativeXml {
	return &TNativeXml{XmlString: "",
		XmlFormat:      xfCompact,
		IndentString:   "  ",
		UseFullNodes:   true,
		RootNodes:      make(map[TXmlElementType]*TXmlNode),
		ParserWarnings: true,
	}
}
func ReadOpenTag(AReader *TsdSurplusReader) (idx int) {
	//Try to read the type of open tag from s
	var Surplus string
	idx = cTagCount - 1
	Candidates := make([]bool, cTagCount)
	for i, _ := range Candidates {
		Candidates[i] = true
	}
	for Found, AIndex := true, 1; Found; AIndex++ {
		Found = false
		if Ch, i := AReader.ReadChar(); i == 0 {
			return idx
		} else {
			Surplus = Surplus + string(Ch)
			for i := cTagCount - 1; i >= 0; i-- {
				if Candidates[i] && (len(cTags[i].FStart) >= AIndex+1) {
					if cTags[i].FStart[AIndex] == Ch {
						Found = true
						if len(cTags[i].FStart) == AIndex+1 {
							idx = i
							break
						}
					} else {
						Candidates[i] = false
					}
				}
			}
		}
	}
	//The surplus string that we already read (everything after the tag)
	AReader.Surplus = Surplus[len(cTags[idx].FStart)-1:]
	return idx
}
func ReadStringFromStreamUntil(AReader *TsdSurplusReader, ASearch string, SkipQuotes bool) (AValue string, b bool) {
	b = false
	InQuotes := false
	//Get last searchstring character
	AIndex := len(ASearch)
	if AIndex == 0 {
		return "", b
	}
	LastSearchChar := ASearch[AIndex-1]
	AValue = ""
	var (
		Ch        byte
		i         int
		QuoteChar byte
	)
	for !b {
		//Add characters to the value to be returned
		if Ch, i = AReader.ReadChar(); i == 0 {
			return AValue, b
		}
		AValue += string(Ch)
		//Do we skip quotes?
		if SkipQuotes {
			if InQuotes && Ch == QuoteChar {
				InQuotes = false
			} else {
				if strings.Index(cQuoteChars, string(Ch)) >= 0 {
					InQuotes = true
					QuoteChar = Ch
				}
			}
		}
		//In quotes? If so ,we don't check the end condition
		if !InQuotes {
			// Is the last char the same as the last char of the search string?
			if Ch == LastSearchChar {
				//Check to see if the whole search string is present
				ValueIndex := len(AValue) - 1
				SearchIndex := len(ASearch) - 1
				if ValueIndex < SearchIndex {
					continue
				}
				b = true
				for SearchIndex > 0 && b {
					b = AValue[ValueIndex] == ASearch[SearchIndex]
					ValueIndex--
					SearchIndex--
				}
			}
		}
	}
	//Use only the part before the search string
	AValue = AValue[:len(AValue)-len(ASearch)]
	return AValue, b
}
func TrimPos(AValue string, Start, Close int) (rStart, rClose int, b bool) {
	//Trim the string in AValue in [Start,Close-1] by adjusting Start and Close variables
	//Checks
	if Start < 0 {
		Start = 0
	}
	if Close < len(AValue)-1 {
		Close = len(AValue) - 1
	}
	if Close <= Start {
		return -1, -1, false
	}
	var bleft, bright bool
	//Trim left,right
	for rStart, rClose = Start, Close; rStart < rClose; {
		//Trim left
		bleft = false
		if strings.IndexAny(string(AValue[rStart]), cControlChars) >= 0 {
			rStart++
			bleft = true
		}
		//Trim right
		bright = false
		if strings.IndexAny(string(AValue[rClose]), cControlChars) >= 0 {
			rClose--
			bright = true
		}
		if !bleft && !bright {
			break
		}
	}
	return rStart, rClose, rClose > rStart
}
func ParseAttributes(AValue string, Start, Close int, Attributes map[string]string) {
	//Convert the attributes string AValue in [Start,Close-1] to the attributes stirnglist
	InQuotes := false
	var AQuoteChar byte = '"'
	if Attributes == nil {
		return
	}
	b := false
	if Start, Close, b = TrimPos(AValue, Start, Close); !b {
		return
	}
	//Clear first
	for k, _ := range Attributes {
		delete(Attributes, k)
	}
	//Loop through characters
	for i := Start; i <= Close; i++ {
		//In quotes?
		if InQuotes {
			if AValue[i] == AQuoteChar {
				InQuotes = false
			}
		} else {
			if strings.Index(cQuoteChars, string(AValue[i])) >= 0 {
				InQuotes = true
				AQuoteChar = AValue[i]
			}
		}
		//Add attribute strings on each controlchar break
		if !InQuotes {
			if strings.Index(cControlChars, string(AValue[i])) >= 0 {
				if i > Start {
					cutstr := string(AValue[Start:i])
					cutstr = strings.Replace(cutstr, string(AQuoteChar), "", len(cutstr))
					if p := strings.Index(cutstr, "="); p > 0 {
						Attributes[cutstr[:p]] = cutstr[p+1:]
					}
				}
				Start = i + 1
			}
		}
	}
	//Add last attribute string
	if Start < Close {
		cutstr := string(AValue[Start:])
		cutstr = strings.Replace(cutstr, string(AQuoteChar), "", len(cutstr))
		if p := strings.Index(cutstr, "="); p > 0 {
			Attributes[cutstr[:p]] = cutstr[p+1:]
		}
	}

	return
}
func ReadStringFromStreamWithQuotes(AReader *TsdSurplusReader, Terminator string) (AValue string, bret bool) {
	AValue = ""
	QuoteChar := byte(0x00)
	InQuotes := false
	var (
		Ch      byte
		readlen int
	)
	for {
		if Ch, readlen = AReader.ReadChar(); readlen != 1 {
			return AValue, false
		}
		if !InQuotes {
			if Ch == '"' || Ch == '\'' {
				InQuotes = true
				QuoteChar = Ch
			} else {
				if Ch == QuoteChar {
					InQuotes = false
				}
			}
		}
		if !InQuotes && string(Ch) == Terminator {
			break
		}
		AValue += string(Ch)
	}
	return AValue, true
}
func WriteStringToStream(S *bytes.Buffer, AString string) {
	if len(AString) > 0 {
		S.WriteString(AString)
	}
}
