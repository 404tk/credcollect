package utils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

func TimeStampFormat(stamp int64) time.Time {
	s1 := time.Unix(stamp, 0)
	if s1.Local().Year() > 9999 {
		return time.Date(9999, 12, 13, 23, 59, 59, 0, time.Local)
	}
	return s1
}

func TimeEpochFormat(epoch int64) time.Time {
	maxTime := int64(99633311740000000)
	if epoch > maxTime {
		return time.Date(2049, 1, 1, 1, 1, 1, 1, time.Local)
	}
	t := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	d := time.Duration(epoch)
	for i := 0; i < 1000; i++ {
		t = t.Add(d)
	}
	return t
}

func ReadFile(filename string) (string, error) {
	s, err := ioutil.ReadFile(filename)
	return string(s), err
}

func GetItemPath(profilePath, file string) (string, error) {
	p, err := filepath.Glob(filepath.Join(profilePath, file))
	if err != nil {
		return "", err
	}
	if len(p) > 0 {
		return p[0], nil
	}
	return "", fmt.Errorf("find %s failed", file)
}
