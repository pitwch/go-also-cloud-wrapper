package alsocloud

import "encoding/xml"

type CIsSessionExpired struct {
	XMLName xml.Name `xml:"IsSessionExpired,omitempty" json:"IsSessionExpired,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}

type CIssueToken struct {
	XMLName xml.Name `xml:"IssueToken,omitempty" json:"IssueToken,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}

type CMessage struct {
	XMLName xml.Name `xml:"Message,omitempty" json:"Message,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}

type CServiceException struct {
	XMLName           xml.Name           `xml:"ServiceException,omitempty" json:"ServiceException,omitempty"`
	AttrXmlnsi        string             `xml:"xmlns i,attr"  json:",omitempty"`
	Attrxmlns         string             `xml:"xmlns,attr"  json:",omitempty"`
	CIsSessionExpired *CIsSessionExpired `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API IsSessionExpired,omitempty" json:"IsSessionExpired,omitempty"`
	CIssueToken       *CIssueToken       `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API IssueToken,omitempty" json:"IssueToken,omitempty"`
	CMessage          *CMessage          `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API Message,omitempty" json:"Message,omitempty"`
}

type CCode struct {
	XMLName xml.Name `xml:"Code,omitempty" json:"Code,omitempty"`
	CValue  *CValue  `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Value,omitempty" json:"Value,omitempty"`
}

type CDetail struct {
	XMLName           xml.Name           `xml:"Detail,omitempty" json:"Detail,omitempty"`
	CServiceException *CServiceException `xml:"http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API ServiceException,omitempty" json:"ServiceException,omitempty"`
}

type CFault struct {
	XMLName   xml.Name `xml:"Fault,omitempty" json:"Fault,omitempty"`
	Attrxmlns string   `xml:"xmlns,attr"  json:",omitempty"`
	CCode     *CCode   `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Code,omitempty" json:"Code,omitempty"`
	CDetail   *CDetail `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Detail,omitempty" json:"Detail,omitempty"`
	CReason   *CReason `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Reason,omitempty" json:"Reason,omitempty"`
}

type CReason struct {
	XMLName xml.Name `xml:"Reason,omitempty" json:"Reason,omitempty"`
	CText   *CText   `xml:"http://schemas.microsoft.com/ws/2005/05/envelope/none Text,omitempty" json:"Text,omitempty"`
}

type CText struct {
	XMLName                                                                                  xml.Name `xml:"Text,omitempty" json:"Text,omitempty"`
	AttrHttp_colon__slash__slash_www_dot_w3_dot_org_slash_XML_slash_1998_slash_namespacelang string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr"  json:",omitempty"`
	String                                                                                   string   `xml:",chardata" json:",omitempty"`
}

type CValue struct {
	XMLName xml.Name `xml:"Value,omitempty" json:"Value,omitempty"`
	String  string   `xml:",chardata" json:",omitempty"`
}
