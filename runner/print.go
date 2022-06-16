package runner

import (
	"fmt"
	"io/ioutil"

	"github.com/modood/table"
)

func (opt *Options) PrintResult(res Output) {
	var content string
	content += check(len(res.Browser), "Browser", res.Browser)
	content += check(len(res.Navicat), "Navicat", res.Navicat)
	content += check(len(res.FileZilla), "FileZilla", res.FileZilla)
	content += check(len(res.WinScp), "WinScp", res.WinScp)
	content += check(len(res.SeeyonOA), "Seeyon", res.SeeyonOA)
	content += check(len(res.DockerHub), "Docker Hub", res.DockerHub)
	if !opt.Silent {
		fmt.Printf(content)
	}
	if opt.Output != "" {
		ioutil.WriteFile(opt.Output, []byte(content), 0664)
	}
}

func check(len int, name string, data interface{}) (c string) {
	if len > 0 {
		c = fmt.Sprintf("%s credentials:\n%s\n", name, table.Table(data))
	}
	return
}
