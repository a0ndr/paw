package pkg

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path"
	"strings"
)

func (dl DefList) Load() error {
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

		def := &Definition{}
		_, err = toml.Decode(string(defFile), def)
		if err != nil {
			return err
		}

		fqn := fmt.Sprintf("%s-%s", def.Name, def.Version)
		if packageDir.Name() != fqn {
			return fmt.Errorf("package directory '%s' does not match it's name and version '%s-%s'", packageDir.Name(), def.Name, def.Version)
		}

		dl[fqn] = def
	}

	return nil
}

func LoadMarks() (map[string]string, error) {
	marks := make(map[string]string)

	markFiles, err := os.ReadDir(Cfg.PackageDir)
	if err != nil {
		return nil, err
	}

	for _, markFile := range markFiles {
		if markFile.IsDir() || !strings.HasSuffix(markFile.Name(), ".mark") {
			continue
		}

		mark, err := os.ReadFile(path.Join(Cfg.PackageDir, markFile.Name()))
		if err != nil {
			return nil, err
		}

		marks[strings.TrimSuffix(markFile.Name(), ".mark")] = string(mark)
	}

	return marks, nil
}
