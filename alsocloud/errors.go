package alsocloud

import "encoding/xml"

type isSessionExpired struct {
	XMLName xml.Name `xml:"IsSessionExpired,omitempty" json:"IsSessionExpired,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}

type issueToken struct {
	XMLName xml.Name `xml:"IssueToken,omitempty" json:"IssueToken,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}

type message struct {
	XMLName xml.Name `xml:"Message,omitempty" json:"Message,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}

type serviceException struct {
	XMLName          xml.Name          `xml:"ServiceException,omitempty" json:"ServiceException,omitempty"`
	AttrXmlnsi       string            `xml:"xmlns i,attr"  json:",omitempty"`
	Attrxmlns        string            `xml:"xmlns,attr"  json:",omitempty"`
	IsSessionExpired *isSessionExpired `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API IsSessionExpired,omitempty" json:"IsSessionExpired,omitempty"`
	IssueToken       *issueToken       `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API IssueToken,omitempty" json:"IssueToken,omitempty"`
	Message          *message          `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API Message,omitempty" json:"Message,omitempty"`
}

type code struct {
	XMLName xml.Name `xml:"Code,omitempty" json:"Code,omitempty"`
	Value   *value   `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Value,omitempty" json:"Value,omitempty"`
}

type detail struct {
	XMLName          xml.Name          `xml:"Detail,omitempty" json:"Detail,omitempty"`
	ServiceException *serviceException `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API ServiceException,omitempty" json:"ServiceException,omitempty"`
}

type cloudfault struct {
	XMLName   xml.Name `xml:"Fault,omitempty" json:"Fault,omitempty"`
	Attrxmlns string   `xml:"xmlns,attr"  json:",omitempty"`
	Code      *code    `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Code,omitempty" json:"Code,omitempty"`
	Detail    *detail  `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Detail,omitempty" json:"Detail,omitempty"`
	Reason    *reason  `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Reason,omitempty" json:"Reason,omitempty"`
}

type reason struct {
	XMLName xml.Name `xml:"Reason,omitempty" json:"Reason,omitempty"`
	CText   *text    `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Text,omitempty" json:"Text,omitempty"`
}

type text struct {
	XMLName  xml.Name `xml:"Text,omitempty" json:"Text,omitempty"`
	AttrHttp string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr"  json:",omitempty"`
	String   string   `xml:",chardata" json:",omitempty"`
}

type value struct {
	XMLName xml.Name `xml:"Value,omitempty" json:"Value,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}
