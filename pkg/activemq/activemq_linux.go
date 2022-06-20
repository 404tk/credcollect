package activemq

import (
	"io/ioutil"
	"os"
	"regexp"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/common/utils"
)

func ActiveMQConsole() []common.ActiveMQPassWord {
	ret := []common.ActiveMQPassWord{}
	filename, err := utils.GetItemPath("/*/*activemq/conf/", "jetty-realm.properties")
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
			if len(m) == 3 {
				ret = append(ret, common.ActiveMQPassWord{
					UserName: m["user"],
					PassWord: m["pass"],
					Role:     m["role"],
				})
			}
		}
	}
	return ret
}

func getParams(content string) ([]map[string]string, error) {
	res := []map[string]string{}
	r, err := regexp.Compile(`(?P<user>\w+):\s?(?P<pass>[\w]+?),\s?(?P<role>\w{3,8})`)
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
