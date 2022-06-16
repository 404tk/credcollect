package navicat

import (
	"log"
	"strconv"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/pkg/navicat/decrypt3"
	"golang.org/x/sys/windows/registry"
)

func Navicat() []common.NavicatPassWord {
	ret := []common.NavicatPassWord{}
	ServersTypes := map[string]string{
		"MySQL Server":      "Software\\PremiumSoft\\Navicat\\Servers",
		"MariaDB Server":    "Software\\PremiumSoft\\NavicatMARIADB\\Servers",
		"MongoDB Server":    "Software\\PremiumSoft\\NavicatMONGODB\\Servers",
		"MSSQL Server":      "Software\\PremiumSoft\\NavicatMSSQL\\Servers",
		"OracleSQL Server":  "Software\\PremiumSoft\\NavicatOra\\Servers",
		"PostgreSQL Server": "Software\\PremiumSoft\\NavicatPG\\Servers",
		"SQLite Server":     "Software\\PremiumSoft\\NavicatSQLite\\Servers",
	}
	for ServersTypeName, ServersRegistryPath := range ServersTypes {
		//fmt.Println("+--------------------------------------------------+")
		//fmt.Println(ServersTypeName)
		//fmt.Println("+--------------------------------------------------+")
		//fmt.Println(ServersRegistryPath)

		key, err := registry.OpenKey(registry.CURRENT_USER, ServersRegistryPath, registry.ALL_ACCESS)
		if err != nil {
			//fmt.Println(err)
			//fmt.Println("No servers is found.")
			continue
		}
		//key.Close()
		kns, err := key.ReadSubKeyNames(0)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, b := range kns {
			key1, err := registry.OpenKey(key, b, registry.ALL_ACCESS)
			if err != nil {
				//fmt.Println("No key")
				log.Println(err)
				continue
			}
			//fmt.Println("Connection Name：", b)
			h, _, err := key1.GetStringValue("Host")
			if err != nil {
				log.Println(err)
				continue
			}
			p, _, err := key1.GetIntegerValue("Port")
			if err != nil {
				log.Println(err)
				continue
			}

			u, _, err := key1.GetStringValue("Username")
			if err != nil {
				log.Println(err)
				continue
			}

			pwd, _, err := key1.GetStringValue("Pwd")
			if err != nil {
				log.Println(err)
				continue
			}
			ret = append(ret, common.NavicatPassWord{
				DbType:         ServersTypeName,
				ConnectionName: b,
				Host:           h,
				Port:           strconv.Itoa(int(p)),
				Username:       u,
				Pwd:            decrypt3.DecryptString(pwd),
			})
			key1.Close()
		}
		key.Close()

	}
	return ret

}
