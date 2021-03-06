package tomcat

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/common/utils"
)

func TomcatManager() []common.TomcatPassWord {
	ret := []common.TomcatPassWord{}
	path := os.Getenv("CATALINA_HOME")
	if path == "" {
		return ret
	}
	filename, err := utils.GetItemPath(filepath.Join(path, "/conf/"), "tomcat-users.xml")
	if err != nil {
		return ret
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
	ps, err := getParams(string(data))
	if err != nil {
		return ret
	}
	if len(ps) > 0 {
		for _, m := range ps {
			if len(m) == 2 {
				ret = append(ret, common.TomcatPassWord{
					UserName: m["user"],
					PassWord: m["pass"],
				})
			}
		}
	}
	return ret
}

func getParams(content string) ([]map[string]string, error) {
	res := []map[string]string{}
	r, err := regexp.Compile(`<user username="(?P<user>\w+)" password="(?P<pass>[^<>]+?)"`)
	if err != nil {
		return res, err
	}
	result := r.FindAllStringSubmatch(content, -1)
	names := r.SubexpNames()
	if len(result) > 0 && len(names) > 1 {
		for _, cred := range result {
			paramsMap := make(map[string]string)
			for i, name := range names {
				if i > 0 {
					paramsMap[name] = cred[i]
				}
			}
			res = append(res, paramsMap)
		}
	}
	return res, nil
}
