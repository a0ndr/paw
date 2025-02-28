package pkg

import (
	"github.com/BurntSushi/toml"
	"github.com/a0ndr/paw/internal"
	"os"
	"sync"
)

type Config struct {
	Debug        bool `toml:"-"`
	DataDir      string
	PackageDir   string
	Repositories map[string]string

	Packages *DefList
	Receipts []*Receipt
	Cache    *Cache
}

var Cfg *Config
var Log *internal.Logger
var configOnce sync.Once

func LoadConfig(path string, debug bool) {
	configOnce.Do(func() {
		var conf Config

		if path == "" {
			path = "/etc/paw.toml"
		}

		Log = &internal.Logger{}
		conf.Debug = debug
		Log.Debug = conf.Debug

		Log.Debugf("Debug: loading config from %s", path)

		content, err := os.ReadFile(path)
		if err != nil {
			Log.Fatalf(ERR_FILE_NOT_FOUND, "Fatal: failed to read config\n")
		}

		_, err = toml.Decode(string(content), &conf)
		if err != nil {
			Log.Fatalf(ERR_FILE_NOT_FOUND, "Fatal: failed to parse config\n")
		}

		Cfg = &conf

		conf.Cache = &Cache{}
		Log.Debugf("Debug: loading package cache")
		if err := conf.Cache.Load(); err != nil {
			Log.Fatalf(ERR_GENERAL, "Could not load cache: %v\n", err)
		}

		conf.Packages = &DefList{}
		Log.Debugf("Debug: loading installed packages")
		if err := conf.Packages.Load(); err != nil {
			Log.Fatalf(ERR_GENERAL, "Could not load packages: %v\n", err)
		}

		Log.Debugf("Debug: loading receipts")
		if err = conf.LoadReceipts(); err != nil {
			Log.Fatalf(ERR_FILE_NOT_FOUND, "Fatal: failed to load receipts: %v\n", err)
		}
	})
}
