package docker

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/common/utils"
	"github.com/tidwall/gjson"
)

func DockerHub() []common.DockerHubPassWord {
	ret := []common.DockerHubPassWord{}
	filename, err := utils.GetItemPath(filepath.Join(os.Getenv("HOME"), "/.docker/"), "config.json")
	if err != nil {
		return ret
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return ret
	}
	h := gjson.GetBytes(content, "auths")
	if h.Exists() {
		for k, v := range h.Map() {
			a := v.Get("auth")
			if a.Exists() {
				u, err := base64.StdEncoding.DecodeString(a.String())
				if err != nil {
					return ret
				}
				account := strings.Split(string(u), ":")
				if len(account) == 2 {
					ret = append(ret, common.DockerHubPassWord{
						Hub:      k,
						UserName: account[0],
						PassWord: account[1],
					})
				}
			}
		}
	}
	return ret
}
