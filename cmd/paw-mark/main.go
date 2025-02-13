package main

import (
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"os"
	"path"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	ToInstall bool `flag:"install" short:"i" help:"Mark package to install"`
	ToRemove  bool `flag:"remove" short:"r" help:"Mark package to remove"`
	Unmark    bool `flag:"unmark" short:"u" help:"Unmark package"`

	Package string `arg:"" help:"Package to mark"`
}

var packages _p.DefList
var cache *_p.Cache

func addMarkfile(mark string, desc string) {
	if mark == "i" {
		_, ok := (*cache.Packages)[CLI.Package]
		if !ok {
			_p.Log.Fatalf(_p.ERR_NOT_FOUND, "Package %s not found\n", CLI.Package)
		}
	} else {
		_, ok := packages[CLI.Package]
		if !ok {
			_p.Log.Fatalf(_p.ERR_NOT_FOUND, "Package %s not installed\n", CLI.Package)
		}
	}

	p := path.Join(_p.Cfg.PackageDir, CLI.Package+".mark")
	err := os.WriteFile(p, []byte(mark), 0644)
	if err != nil {
		_p.Log.Fatalf(_p.ERR_OS_ERROR, "Fatal: could not open file for writing: %v\n", err)
	}

	_p.Log.Printf("Package '%s' marked for %s\n", CLI.Package, desc)
}

func removeMarkfile(mark string, desc1 string, desc2 string) {
	p := path.Join(_p.Cfg.PackageDir, CLI.Package+".mark")
	err := os.Remove(p)
	if err != nil {
		_p.Log.Fatalf(_p.ERR_NOT_FOUND, "Package \"%s\" is not marked for %s\n", CLI.Package, desc1)
	}

	_p.Log.Printf("Package '%s' won't be %s\n", CLI.Package, desc2)
}

func removeAllMarkfiles() {
	p := path.Join(_p.Cfg.PackageDir, CLI.Package+".mark")
	err := os.Remove(p)
	if err != nil {
		p = path.Join(_p.Cfg.PackageDir, CLI.Package+".mark")
		err = os.Remove(p)
		if err != nil {
			_p.Log.Fatalf(_p.ERR_NOT_FOUND, "Package \"%s\" was not marked\n", CLI.Package)
		}
	}

	_p.Log.Printf("Package '%s' unmarked\n", CLI.Package)
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)

	packages = _p.DefList{}
	if err := packages.Load(); err != nil {
		_p.Log.Fatalf(_p.ERR_GENERAL, "Could not load packages: %v\n", err)
	}

	cache = &_p.Cache{Packages: &_p.DefList{}}
	err := cache.Load()
	if err != nil {
		_p.Log.Fatalf(_p.ERR_GENERAL, "Failed to load cache: %v\n", err)
	}

	switch {
	case CLI.ToInstall && !CLI.Unmark:
		addMarkfile("i", "installation")
	case CLI.ToRemove && !CLI.Unmark:
		addMarkfile("r", "removal")
	case CLI.ToInstall && CLI.Unmark:
		removeMarkfile("i", "installation", "installed")
	case CLI.ToRemove && CLI.Unmark:
		removeMarkfile("r", "removal", "removed")
	case CLI.Unmark && !CLI.ToInstall && !CLI.ToRemove:
		removeAllMarkfiles()
	default:
		_p.Log.Fatalf(_p.ERR_INVALID_PARAMS, "Please specify either --install, --remove, or --unmark and a package to modify\n")
	}
}
