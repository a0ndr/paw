package pkg

import (
	"github.com/BurntSushi/toml"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Package struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Mark        string `toml:"-"`
	Description string `toml:"description"`
}

type TPackages []*Package

var loadPackagesOnce sync.Once

func Packages() TPackages {
	loadPackagesOnce.Do(func() {
		entries, err := os.ReadDir(Cfg.DataDir)
		if err != nil {
			Log.Fatalf(ERR_OS_ERROR, "Fatal: failed to read data directory\n")
		}

		for _, entry := range entries {
			if entry.IsDir() {
				Log.Debugf("Debug: reading directory \"%s\"", entry.Name())

				if len(strings.Split(entry.Name(), "-")) < 2 {
					Log.Errorf("Warning: invalid package directory name \"%s\", skipping\n", entry.Name())
					continue
				}

				name := strings.Split(entry.Name(), "-")[0]
				version := strings.SplitN(entry.Name(), ":", 1)[1]

				Log.Debugf("Debug: found package \"%s\" version %s", name, version)

				metaPath := filepath.Join(Cfg.DataDir, entry.Name(), "PKGMETA")
				meta, err := os.ReadFile(metaPath)
				if err != nil {
					Log.Errorf("Warning: failed to read PKGMETA of package \"%s\"\n", entry.Name())
					continue
				}

				var pkg Package
				_, err = toml.Decode(string(meta), &pkg)
				if err != nil {
					Log.Errorf("Warning: failed to decode PKGMETA of package \"%s\"\n", entry.Name())
					continue
				}

				if pkg.Name != name {
					Log.Errorf("Warning: package \"%s\" name mismatch, did you manually edit the PKGMETA file?\n", entry.Name())
					continue
				}
				if pkg.Version != version {
					Log.Errorf("Warning: package \"%s\" version mismatch, did you manually edit the PKGMETA file?\n", entry.Name())
					continue
				}

				Cfg.packages = append(Cfg.packages, &pkg)
				continue
			}
			if strings.HasSuffix(entry.Name(), "#i") || strings.HasSuffix(entry.Name(), "#r") {
				Log.Debugf("Debug: found markfile \"%s\"", entry.Name())

				p := path.Join(Cfg.DataDir, entry.Name())
				content, err := os.ReadFile(p)
				if err != nil {
					Log.Errorf("Warning: failed to read markfile \"%s\"\n", entry.Name())
					continue
				}

				var pkg Package
				_, err = toml.Decode(string(content), &pkg)
				if err != nil {
					Log.Errorf("Warning: failed to decode markfile \"%s\"\n", entry.Name())
					continue
				}

				found := false
				for _, existing := range Cfg.packages {
					if existing.Name == pkg.Name {
						existing.Mark = strings.Split(entry.Name(), "#")[1]
						found = true
						break
					}
				}

				if found {
					continue
				}

				pkg.Mark = strings.Split(entry.Name(), "#")[1]
				Cfg.packages = append(Cfg.packages, &pkg)

				continue
			}

			Log.Debugf("unknown file \"%s\", skipping", entry.Name())
		}
	})

	return Cfg.packages
}

func (pkgs TPackages) FindFQN(fqn string) *Package {
	for _, pkg := range pkgs {
		if pkg.Name+":"+pkg.Version == fqn {
			return pkg
		}
	}
	return nil
}
