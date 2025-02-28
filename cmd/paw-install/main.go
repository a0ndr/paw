package main

import (
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	Packages []string `arg:"" help:"Packages to install" required:""`
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)

	cfg := _p.Cfg
	log := _p.Log
	//packages := cfg.Packages
	cache := cfg.Cache

	receipt := _p.CreateReceipt()

	for _, pkgName := range CLI.Packages {
		pkg, ok := (*cache.Packages)[pkgName]
		if !ok {
			log.Fatalf(_p.ERR_NOT_FOUND, "Error: package %s not found\n", pkgName)
		}

		for _, dependency := range pkg.FindDependencies() {
			receipt.AddOperation(dependency, "install")
		}

		receipt.AddOperation(pkg, "install")
	}

	err := receipt.Flush()
	if err != nil {
		log.Fatalf(_p.ERR_OS_ERROR, "Error: failed to flush receipt: %v", err)
	}
}
