package pkg

import (
	"github.com/BurntSushi/toml"
	"github.com/a0ndr/paw/internal"
	"os"
	"sync"
)

type Config struct {
	Debug        bool
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

func LoadConfig(path string) {
	configOnce.Do(func() {
		var conf Config

		if path == "" {
			path = "/etc/paw.toml"
		}

		content, err := os.ReadFile(path)
		if err != nil {
			Log.Fatalf(ERR_FILE_NOT_FOUND, "Fatal: failed to read config\n")
		}

		_, err = toml.Decode(string(content), &conf)
		if err != nil {
			Log.Fatalf(ERR_FILE_NOT_FOUND, "Fatal: failed to parse config\n")
		}

		Log = &internal.Logger{}
		Log.Debug = conf.Debug

		Cfg = &conf

		conf.Cache = &Cache{}
		if err := conf.Cache.Load(); err != nil {
			Log.Fatalf(ERR_GENERAL, "Could not load cache: %v\n", err)
		}

		conf.Packages = &DefList{}
		if err := conf.Packages.Load(); err != nil {
			Log.Fatalf(ERR_GENERAL, "Could not load packages: %v\n", err)
		}

		if err = conf.LoadReceipts(); err != nil {
			Log.Fatalf(ERR_FILE_NOT_FOUND, "Fatal: failed to load receipts: %v\n", err)
		}
	})
}
