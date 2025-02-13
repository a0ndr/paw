package main

import (
	"fmt"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"strings"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	LocalOnly        bool `flag:"local" short:"L" help:"List only installed packages"`
	WithDescriptions bool `flag:"" short:"d" help:"Show descriptions"`
	WithMarks        bool `flag:"" short:"m" help:"Show marks"`

	ToInstall bool `flag:"install" short:"i" help:"Lists marked as to install"`
	ToRemove  bool `flag:"remove" short:"r" help:"Lists marked as to remove"`
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)

	cache := _p.Cache{}
	if err := cache.Load(); err != nil {
		_p.Log.Fatalf(_p.ERR_GENERAL, "Could not load cache: %v\n", err)
	}

	packages := _p.DefList{}
	if err := packages.Load(); err != nil {
		_p.Log.Fatalf(_p.ERR_GENERAL, "Could not load packages: %v\n", err)
	}

	if CLI.ToInstall || CLI.ToRemove {
		marks, err := _p.LoadMarks()
		if err != nil {
			_p.Log.Fatalf(_p.ERR_GENERAL, "Could not load marks: %v\n", err)
		}

		displayPackages := _p.DefList{}

		if CLI.ToInstall {
			for pkg, mark := range marks {
				pkg = strings.TrimSuffix(pkg, ".mark")
				if mark == "i" {
					def, ok := (*cache.Packages)[pkg]
					if !ok {
						_p.Log.Errorf("Warning: install mark for nonexistent package %s\n", pkg)
						continue
					}

					displayPackages[pkg] = def
				}
			}
		}

		if CLI.ToRemove {
			for pkg, mark := range marks {
				pkg = strings.TrimSuffix(pkg, ".mark")
				if mark == "r" {
					def, ok := packages[pkg]
					if !ok {
						_p.Log.Errorf("Warning: remove mark for not installed package %s\n", pkg)
						continue
					}

					displayPackages[pkg] = def
				}
			}
		}

		for pkg, def := range displayPackages {
			description := ""
			if CLI.WithDescriptions {
				description = "  " + def.Description
			}

			mark := ""
			if CLI.WithMarks {
				mark = " [" + marks[pkg] + "]"
			}

			fmt.Printf("%s%s%s\n", pkg, description, mark)
		}

		return
	}

	if !CLI.LocalOnly {
		for pkg, def := range *cache.Packages {
			packages[pkg] = def
		}
	}

	for pkg, def := range packages {
		description := ""
		if CLI.WithDescriptions {
			description = "  " + def.Description
		}

		fmt.Printf("%s%s\n", pkg, description)
	}
}
