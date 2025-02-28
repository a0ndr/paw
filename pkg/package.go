package pkg

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gobwas/glob"
	"os"
	"path"
	"strings"
)

func (dl *DefList) Load() error {
	packageDirs, err := os.ReadDir(Cfg.PackageDir)
	if err != nil {
		return err
	}

	for _, packageDir := range packageDirs {
		if !packageDir.IsDir() {
			continue
		}

		defFile, err := os.ReadFile(path.Join(Cfg.PackageDir, packageDir.Name(), "PKG"))
		if err != nil {
			return err
		}

		def := &Package{}
		_, err = toml.Decode(string(defFile), def)
		if err != nil {
			return err
		}

		fqn := fmt.Sprintf("%s-%s", def.Name, def.Version)
		if packageDir.Name() != fqn {
			return fmt.Errorf("package directory '%s' does not match it's name and version '%s-%s'", packageDir.Name(), def.Name, def.Version)
		}

		(*dl)[fqn] = def
	}

	return nil
}

func (cache *Cache) FindLatest(name string) *Package {
	name = name + "-*"
	g := glob.MustCompile(name)

	var latest *Package
	for pkg, def := range *cache.Packages {
		if g.Match(pkg) && (latest == nil || def.BuiltAt.After(latest.BuiltAt)) {
			latest = def
		}
	}

	return latest
}

func (cache *Cache) FindVersions(name string) []*Package {
	name = name + "-*"
	g := glob.MustCompile(name)

	var pkgs []*Package
	for pkg, def := range *cache.Packages {
		if g.Match(pkg) {
			pkgs = append(pkgs, def)
		}
	}

	return pkgs
}

func (dl *DefList) Find(name string) *Package {
	g := glob.MustCompile(name)

	for pkg, def := range *dl {
		if g.Match(pkg) {
			return def
		}
	}

	return nil
}

func (dl *DefList) IsInstalled(name string) bool {
	return dl.Find(name) != nil
}

func (pkg *Package) FindConflicts() []*Package {
	var conflicts []*Package

	for _, pkgConflict := range pkg.Conflicts {
		conflict := Cfg.Packages.Find(pkgConflict)
		if conflict != nil {
			conflicts = append(conflicts, conflict)
		}
	}

	for _, installedPkg := range *Cfg.Packages {
		pkgConflicts := installedPkg.FindConflicts()
		for _, pkgConflict := range pkgConflicts {
			if pkgConflict != nil && pkgConflict.Name == pkg.Name {
				conflicts = append(conflicts, pkgConflict)
			}
		}
	}

	return conflicts
}

func (pkg *Package) FindDependencies() []*Package {
	var dependencies []*Package

	for _, dep := range pkg.Dependencies {
		dependency := Cfg.Cache.FindLatest(dep)
		if dependency != nil {
			conflicts := dependency.FindConflicts()
			conflictNames := make([]string, len(conflicts))
			for i, conflict := range conflicts {
				conflictNames[i] = conflict.Name
			}

			if len(conflicts) > 0 {
				Log.Fatalf(ERR_CONFLICT, "Error: package %s-%s conflicts with %s\n", dependency.Name, dependency.Version, strings.Join(conflictNames, ", "))
			}
			for _, dep2 := range dependency.FindDependencies() {
				dependencies = append(dependencies, dep2)
			}
			dependencies = append(dependencies, dependency)
		} else {
			Log.Fatalf(ERR_NOT_FOUND, "Dependency '%s' of '%s' not found\n", dep, pkg.Name)
		}
	}

	return dependencies
}
