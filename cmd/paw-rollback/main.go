package main

import (
	"github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"os"
	"path"
	"strings"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`
}

func main() {
	_ = kong.Parse(&CLI)
	pkg.LoadConfig(CLI.Config)

	entries, err := os.ReadDir(pkg.Cfg.PackageDir)
	if err != nil {
		pkg.Log.Fatal(pkg.ERR_OS_ERROR, "Fatal: failed to read data directory\n")
	}

	count := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".mark") {
			pkg.Log.Debugf("Debug: found markfile \"%s\"", entry.Name())

			p := path.Join(pkg.Cfg.PackageDir, entry.Name())
			err = os.Remove(p)
			if err != nil {
				pkg.Log.Fatalf(pkg.ERR_OS_ERROR, "Failed: failed to remove markfile \"%s\": %s\n", entry.Name(), err.Error())
			}

			count++
		}
	}

	pkg.Log.Printf("Rolled back %d changes.\n", count)
}
