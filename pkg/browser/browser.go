package browser

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/404tk/credcollect/common/utils"
)

const (
	chromeName         = "Chrome"
	chromeBetaName     = "Chrome Beta"
	chromiumName       = "Chromium"
	edgeName           = "Microsoft Edge"
	firefoxName        = "Firefox"
	firefoxBetaName    = "Firefox Beta"
	firefoxDevName     = "Firefox Dev"
	firefoxNightlyName = "Firefox Nightly"
	firefoxESRName     = "Firefox ESR"
	speed360Name       = "360speed"
	qqBrowserName      = "qq"
	braveName          = "Brave"
	operaName          = "Opera"
	operaGXName        = "OperaGX"
	vivaldiName        = "Vivaldi"
)

type Browser interface {
	// InitSecretKey is init chrome secret key, firefox's key always empty
	InitSecretKey() error

	// GetName return browser name
	GetName() string

	// GetSecretKey return browser secret key
	GetSecretKey() []byte

	// GetItem return password
	GetItem() (Item, error)
}

const (
	password = "password"
)

var (
	errItemNotSupported  = errors.New(`item not supported, default is "all", choose from password`)
	errDbusSecretIsEmpty = errors.New("dbus secret key is empty")
)

var (
	chromiumItems = map[string]struct {
		mainFile string
		newItem  func(mainFile, subFile string) Item
	}{
		password: {
			mainFile: ChromePasswordFile,
			newItem:  NewCPasswords,
		},
	}
	firefoxItems = map[string]struct {
		mainFile string
		subFile  string
		newItem  func(mainFile, subFile string) Item
	}{
		password: {
			mainFile: FirefoxKey4File,
			subFile:  FirefoxLoginFile,
			newItem:  NewFPasswords,
		},
	}
)

type Chromium struct {
	name        string
	profilePath string
	keyPath     string
	storage     string // storage use for linux and macOS, get secret key
	secretKey   []byte
}

// NewChromium return Chromium browser interface
func NewChromium(profile, key, name, storage string) (Browser, error) {
	return &Chromium{profilePath: profile, keyPath: key, name: name, storage: storage}, nil
}

func (c *Chromium) GetName() string {
	return c.name
}

func (c *Chromium) GetSecretKey() []byte {
	return c.secretKey
}

// GetItem return single item
func (c *Chromium) GetItem() (Item, error) {
	if item, ok := chromiumItems[password]; ok {
		m, err := utils.GetItemPath(c.profilePath, item.mainFile)
		if err != nil {
			// log.Printf("%s find %s file failed, ERR:%s\n", c.name, item.mainFile, err.Error())
			return nil, err
		}
		i := item.newItem(m, "")
		return i, nil
	} else {
		return nil, errItemNotSupported
	}
}

type Firefox struct {
	name        string
	profilePath string
	keyPath     string
}

// NewFirefox return firefox browser interface
func NewFirefox(profile, key, name, storage string) (Browser, error) {
	return &Firefox{profilePath: profile, keyPath: key, name: name}, nil
}

func (f *Firefox) GetItem() (Item, error) {
	if item, ok := firefoxItems[password]; ok {
		var (
			sub, main string
			err       error
		)
		if item.subFile != "" {
			sub, err = utils.GetItemPath(f.profilePath, item.subFile)
			if err != nil {
				// log.Printf("%s find %s file failed, ERR:%s\n", f.name, item.subFile, err.Error())
				return nil, err
			}
		}
		main, err = utils.GetItemPath(f.profilePath, item.mainFile)
		if err != nil {
			// log.Printf("%s find %s file failed, ERR:%s\n", f.name, item.mainFile, err.Error())
			return nil, err
		}
		i := item.newItem(main, sub)
		// log.Printf("%s find %s file success\n", f.name, item.mainFile)
		return i, nil
	} else {
		return nil, errItemNotSupported
	}
}

func (f *Firefox) GetName() string {
	return f.name
}

// GetSecretKey for firefox is always nil
// this method used to implement Browser interface
func (f *Firefox) GetSecretKey() []byte {
	return nil
}

// InitSecretKey for firefox is always nil
// this method used to implement Browser interface
func (f *Firefox) InitSecretKey() error {
	return nil
}

// PickBrowser return a list of browser interface
func PickBrowser() (browsers []Browser) {
	for _, v := range browserList {
		b, err := v.New(v.ProfilePath, v.KeyPath, v.Name, v.Storage)
		if err != nil {
			log.Println(err)
		}
		browsers = append(browsers, b)
	}
	return browsers
}

// PickCustomBrowser pick single browser with custom browser profile path and key file path (windows only).
// If custom key file path is empty, but the current browser requires key file (chromium for windows version > 80)
// key file path will be automatically found in the profile path's parent directory.
func PickCustomBrowser(browserName, cusProfile, cusKey string) ([]Browser, error) {
	var (
		browsers []Browser
	)
	browserName = strings.ToLower(browserName)
	supportBrowser := strings.Join(ListBrowser(), "|")
	if browserName == "all" {
		return nil, fmt.Errorf("can't select all browser, pick one from %s with -b flag\n", supportBrowser)
	}
	if choice, ok := browserList[browserName]; ok {
		// if this browser need key path
		if choice.KeyPath != "" {
			var err error
			// if browser need key path and cusKey is empty, try to get key path with profile dir
			if cusKey == "" {
				cusKey, err = getKeyPath(cusProfile)
				if err != nil {
					return nil, err
				}
			}
			if err := checkKeyPath(cusKey); err != nil {
				return nil, err
			}
		}
		b, err := choice.New(cusProfile, cusKey, choice.Name, choice.Storage)
		browsers = append(browsers, b)
		return browsers, err
	} else {
		return nil, fmt.Errorf("%s not support, pick one from %s with -b flag\n", browserName, supportBrowser)
	}
}

// getKeyPath try to get key file path with the browser's profile path
// default key file path is in the parent directory of the profile dir, and name is [Local State]
func getKeyPath(profilePath string) (string, error) {
	if _, err := os.Stat(filepath.Clean(profilePath)); os.IsNotExist(err) {
		return "", err
	}
	parentDir := getParentDirectory(profilePath)
	keyPath := filepath.Join(parentDir, "Local State")
	return keyPath, nil
}

// check key file path is exist
func checkKeyPath(keyPath string) error {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("secret key path not exist, please check %s", keyPath)
	}
	return nil
}

func getParentDirectory(dir string) string {
	var (
		length int
	)
	// filepath.Clean(dir) auto remove
	dir = strings.ReplaceAll(filepath.Clean(dir), `\`, `/`)
	length = strings.LastIndex(dir, "/")
	if length > 0 {
		if length > len([]rune(dir)) {
			length = len([]rune(dir))
		}
		return string([]rune(dir)[:length])
	}
	return ""
}

func ListBrowser() []string {
	var l []string
	for k := range browserList {
		l = append(l, k)
	}
	return l
}

func ListItem() []string {
	var l []string
	for k := range chromiumItems {
		l = append(l, k)
	}
	return l
}
