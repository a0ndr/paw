package main

import (
	"fmt"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	LocalOnly        bool `flag:"local" short:"L" help:"List only installed packages"`
	WithDescriptions bool `flag:"" short:"d" help:"Show descriptions"`
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)
	cfg := _p.Cfg

	if CLI.LocalOnly {
		for pkg, def := range *cfg.Packages {
			description := ""
			if CLI.WithDescriptions {
				description = "  " + def.Description
			}

			fmt.Printf("%s%s\n", pkg, description)
		}
	} else {
		for pkg, def := range *cfg.Cache.Packages {
			description := ""
			if CLI.WithDescriptions {
				description = "  " + def.Description
			}

			fmt.Printf("%s%s\n", pkg, description)
		}
	}
}
