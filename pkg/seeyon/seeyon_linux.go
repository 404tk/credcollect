package seeyon

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/common/utils"
)

func SeeyonOA() (ret []common.SeeyonPassWord) {
	p, err := utils.GetItemPath("/*/Seeyon/*/base/conf/", "datasourceCtp.properties")
	if err != nil {
		return
	}
	file, err := os.Open(p)
	if err != nil {
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	ps, err := getParams(string(data))
	if len(ps) == 5 {
		ret = append(ret, common.SeeyonPassWord{
			DbType:   ps["dbtype"],
			DbName:   ps["dbname"],
			Host:     ps["host"],
			Username: ps["user"],
			Pwd:      dbPwdDecode(ps["pwd"]),
		})
	}
	return
}

func getParams(content string) (map[string]string, error) {
	paramsMap := make(map[string]string)
	r, err := regexp.Compile(`ctpDataSource.username=(?P<user>[\w]+)[\s\S]+password=/[\d.]+/(?P<pwd>[\w=]+)\s+ctpDataSource.url=jdbc:(?P<dbtype>[\w]+)://(?P<host>[\w.:]+);DatabaseName=(?P<dbname>[\w]+)`)
	if err != nil {
		return paramsMap, err
	}
	result := r.FindStringSubmatch(content)
	names := r.SubexpNames()
	if len(result) > 1 && len(names) > 1 {
		for i, name := range names {
			if i > 0 && i <= len(result) {
				paramsMap[name] = result[i]
			}
		}
	}
	return paramsMap, nil
}

func dbPwdDecode(pwd string) string {
	bytes, err := base64.StdEncoding.DecodeString(pwd)
	if err != nil {
		return ""
	}
	var res []byte
	for _, b := range bytes {
		res = append(res, b-1)
	}
	return string(res)
}
