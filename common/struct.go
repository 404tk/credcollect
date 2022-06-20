package common

type BrowserPassword struct {
	BrowserName string `table:"BrowserName"`
	UserName    string `table:"UserName"`
	PassWord    string `table:"Password"`
	LoginUrl    string `table:"LoginUrl"`
	CreateDate  string `table:"CreateDate"`
}

type NavicatPassWord struct {
	DbType         string `table:"DbType"`
	ConnectionName string `table:"ConnectionName"`
	Host           string `table:"Host"`
	Port           string `table:"Port"`
	Username       string `table:"UserName"`
	Pwd            string `table:"Password"`
}

type FileZillaPassWord struct {
	HostName   string `table:"Host"`
	PortNumber string `table:"Port"`
	UserName   string `table:"UserName"`
	PassWord   string `table:"Password"`
}

type WinScpPassWord struct {
	HostName   string `table:"Host"`
	PortNumber string `table:"Port"`
	UserName   string `table:"UserName"`
	PassWord   string `table:"Password"`
}

type SeeyonPassWord struct {
	DbType   string `table:"Type"`
	DbName   string `table:"Database Name"`
	Host     string `table:"Host(:Port)"`
	Username string `table:"UserName"`
	Pwd      string `table:"Password"`
}

type DockerHubPassWord struct {
	Hub      string `table:"Hub Address"`
	UserName string `table:"UserName"`
	PassWord string `table:"Password"`
}

type TomcatPassWord struct {
	UserName string `table:"UserName"`
	PassWord string `table:"Password"`
}

type ActiveMQPassWord struct {
	UserName string `table:"UserName"`
	PassWord string `table:"Password"`
	Role     string `table:"Role"`
}
