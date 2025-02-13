package main

import (
	"github.com/BurntSushi/toml"
	i "github.com/a0ndr/paw/internal"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"io"
	"net/http"
	"os"
	"path"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`
}

func download(url, output string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)

	i.Logf("Syncing repo package lists...")
	cache := &_p.Cache{Packages: &_p.DefList{}}

	for name, repo := range _p.Cfg.Repositories {
		i.Log2f("Syncing %s (%s)...", name, repo)

		resp, err := http.Get(repo + "/PKGLIST")
		if err != nil {
			i.Error2f("Syncing %s (%s): %s", name, repo, err)
			os.Exit(1)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		var list *_p.List
		_, err = toml.Decode(string(body), &list)
		if err != nil {
			i.Error2f("Syncing %s (%s): %s", name, repo, err)
			os.Exit(1)
		}

		for fqn, pkg := range list.Packages {
			if existing, ok := (*cache.Packages)[pkg.Name]; ok {
				i.Log2f("Package %s conflict (repos: %s & %s), overwriting", fqn, name, existing.Repository)
			}

			(*cache.Packages)[fqn] = pkg.ToDefinition(name)
		}
	}

	f, err := os.OpenFile(path.Join(_p.Cfg.DataDir, "cache.toml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		i.Errorf("Failed to open cache: %s", err)
		os.Exit(1)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	encoder := toml.NewEncoder(f)
	if err = encoder.Encode(cache); err != nil {
		i.Errorf("Failed to write cache: %s", err)
		os.Exit(1)
	}
}
