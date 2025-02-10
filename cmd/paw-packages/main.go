package main

import (
	"fmt"
	"github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	WithDescriptions bool `flag:"" short:"d" help:"Show descriptions"`
	WithMarks        bool `flag:"" short:"m" help:"Show marks"`

	ToInstall bool `flag:"install" short:"i" help:"Lists marked as to install"`
	ToRemove  bool `flag:"remove" short:"r" help:"Lists marked as to remove"`
	ToChange  bool `flag:"change" short:"c" help:"Lists marked as to install or remove"`
}

func main() {
	_ = kong.Parse(&CLI)
	pkg.LoadConfig(CLI.Config)

	for _, pak := range pkg.Packages() {
		if (CLI.ToChange && pak.Mark == "") || (CLI.ToInstall && pak.Mark != "i") || (CLI.ToRemove && pak.Mark != "r") {
			continue
		}

		desc := ""
		if CLI.WithDescriptions {
			desc = " " + pak.Description
		}

		mark := ""
		if CLI.WithMarks && pak.Mark != "" {
			mark = fmt.Sprintf(" [%s]", pak.Mark)
		}

		fmt.Printf("%s:%s%s%s\n", pak.Name, pak.Version, desc, mark)
	}
}
