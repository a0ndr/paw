package main

import (
	"github.com/BurntSushi/toml"
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

func addMarkfile(mark string, desc string) {
	p := path.Join(_p.Cfg.DataDir, CLI.Package+"#"+mark)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		_p.Log.Fatalf(_p.ERR_OS_ERROR, "Fatal: could not open file for writing: %v\n", err)
	}

	pkg := _p.Packages().FindFQN(CLI.Package)
	if pkg == nil {
		_p.Log.Fatalf(_p.ERR_NOT_FOUND, "Package '%s' not found\n", CLI.Package)
	}

	decoder := toml.NewEncoder(f)
	err = decoder.Encode(pkg)
	if err != nil {
		_p.Log.Fatalf(_p.ERR_OS_ERROR, "Fatal: failed to decode metadata: %v\n", err)
	}

	_p.Log.Printf("Package '%s' marked for %s\n", CLI.Package, desc)
}

func removeMarkfile(mark string, desc1 string, desc2 string) {
	p := path.Join(_p.Cfg.DataDir, CLI.Package+"#"+mark)
	err := os.Remove(p)
	if err != nil {
		_p.Log.Fatalf(_p.ERR_NOT_FOUND, "Package \"%s\" is not marked for %s\n", CLI.Package, desc1)
	}

	_p.Log.Printf("Package '%s' won't be %s\n", CLI.Package, desc2)
}

func removeAllMarkfiles() {
	p := path.Join(_p.Cfg.DataDir, CLI.Package+"#i")
	err := os.Remove(p)
	if err != nil {
		p = path.Join(_p.Cfg.DataDir, CLI.Package+"#r")
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
