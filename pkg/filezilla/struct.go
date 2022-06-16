package filezilla

type Server struct {
	Host         string `xml:"Host"`
	Port         string `xml:"Port"`
	Protocol     int    `xml:"Protocol"`
	Type         int    `xml:"Type"`
	User         string `xml:"User"`
	Pass         string `xml:"Pass"`
	Logontype    int    `xml:"Logontype"`
	PasvMode     string `xml:"PasvMode"`
	EncodingType string `xml:"EncodingType"`
	BypassProxy  bool   `xml:"BypassProxy"`
}

type RecentServer struct {
	Server []Server `xml:"Server"`
}

type FileZillaXML struct {
	Version       string       `xml:"version,attr"`
	Platform      string       `xml:"platform,attr"`
	RecentServers RecentServer `xml:"RecentServers"`
}
