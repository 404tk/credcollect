package browser

const (
	fireFoxProfilePath        = "/Users/*/Library/Application Support/Firefox/Profiles/*.default-release*/"
	fireFoxBetaProfilePath    = "/Users/*/Library/Application Support/Firefox/Profiles/*.default-beta*/"
	fireFoxDevProfilePath     = "/Users/*/Library/Application Support/Firefox/Profiles/*.dev-edition-default*/"
	fireFoxNightlyProfilePath = "/Users/*/Library/Application Support/Firefox/Profiles/*.default-nightly*/"
	fireFoxESRProfilePath     = "/Users/*/Library/Application Support/Firefox/Profiles/*.default-esr*/"
)

var (
	browserList = map[string]struct {
		ProfilePath string
		Name        string
		KeyPath     string
		Storage     string
		New         func(profile, key, name, storage string) (Browser, error)
	}{
		"firefox": {
			ProfilePath: fireFoxProfilePath,
			Name:        firefoxName,
			New:         NewFirefox,
		},
		"firefox-beta": {
			ProfilePath: fireFoxBetaProfilePath,
			Name:        firefoxBetaName,
			New:         NewFirefox,
		},
		"firefox-dev": {
			ProfilePath: fireFoxDevProfilePath,
			Name:        firefoxDevName,
			New:         NewFirefox,
		},
		"firefox-nightly": {
			ProfilePath: fireFoxNightlyProfilePath,
			Name:        firefoxNightlyName,
			New:         NewFirefox,
		},
		"firefox-esr": {
			ProfilePath: fireFoxESRProfilePath,
			Name:        firefoxESRName,
			New:         NewFirefox,
		},
	}
)

func (c *Chromium) InitSecretKey() error {
	return nil
}
