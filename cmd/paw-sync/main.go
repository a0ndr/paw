package main

import (
	i "github.com/a0ndr/paw/internal"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)

	i.Logf("Syncing mirror package lists...")

	for name, mirror := range _p.Cfg.Mirrors {
		i.Log2f("Syncing %s (%s)...", name, mirror)
	}
}
