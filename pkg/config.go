package pkg

import (
	"github.com/BurntSushi/toml"
	"github.com/a0ndr/paw/internal"
	"os"
	"sync"
)

type Config struct {
	Debug   bool
	DataDir string
	Mirrors map[string]string

	packages TPackages
}

var Cfg *Config
var Log internal.Logger
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

		Log.Debug = conf.Debug

		Cfg = &conf
	})
}
