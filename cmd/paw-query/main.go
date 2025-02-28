package main

import (
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"strings"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	Debug bool `flag:"" help:"Enable debug mode"`

	Package string `arg:"" help:"Package name" required:""`
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config, CLI.Debug)

	cfg := _p.Cfg
	log := _p.Log
	pkgs := cfg.Packages
	cache := cfg.Cache

	pkg := pkgs.Find(CLI.Package)
	if pkg != nil {
		log.Printf("Name: %s\n", pkg.Name)
		log.Printf("Installed version: %s\n", pkg.Version)
		log.Printf("Description: %s\n", pkg.Description)
		if len(pkg.Dependencies) > 0 {
			log.Printf("Dependencies: \n    %s\n", strings.Join(pkg.Dependencies, "\n    "))
		}
		if len(pkg.SoftDependencies) > 0 {
			log.Printf("Soft dependencies: \n    %s\n", strings.Join(pkg.SoftDependencies, "\n    "))
		}
		if len(pkg.Conflicts) > 0 {
			log.Printf("Conflicts: \n    %s\n", strings.Join(pkg.Conflicts, "\n    "))
		}
		return
	}

	pkg = cache.FindLatest(CLI.Package)
	if pkg == nil {
		log.Fatalf(_p.ERR_NOT_FOUND, "Package %s not found", CLI.Package)
	}

	log.Printf("Name: %s\n", pkg.Name)
	log.Printf("Description: %s\n", pkg.Description)
	if len(pkg.Dependencies) > 0 {
		log.Printf("Dependencies: \n    %s\n", strings.Join(pkg.Dependencies, "\n    "))
	}
	if len(pkg.SoftDependencies) > 0 {
		log.Printf("Soft dependencies: \n    %s\n", strings.Join(pkg.SoftDependencies, "\n    "))
	}
	if len(pkg.Conflicts) > 0 {
		log.Printf("Conflicts: \n    %s\n", strings.Join(pkg.Conflicts, "\n    "))
	}
	var versions []string
	for _, _pkg := range cache.FindVersions(CLI.Package) {
		versions = append(versions, _pkg.Version)
	}
	log.Printf("Versions: \n    %s\n", strings.Join(versions, "\n    "))
}
