package main

import (
	"github.com/BurntSushi/toml"
	i "github.com/a0ndr/paw/internal"
	"github.com/alecthomas/kong"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var CLI struct {
	Path string `arg:"" help:"Package to generate checksums for" default:"."`
}

type Pkgbuild struct {
	Name        string `toml:"Name"`
	Version     string `toml:"Version"`
	Description string `toml:"Description"`
	Source      string `toml:"Source"`
	Build       string `toml:"Build"`
	Install     string `toml:"Install"`
	Configure   string `toml:"Configure"`
}

func main() {
	_ = kong.Parse(&CLI)

	stat, err := os.Stat(CLI.Path)
	if err != nil {
		i.Errorf("Failed to stat target path: %s", err.Error())
		os.Exit(1)
	}
	if !stat.IsDir() {
		i.Errorf("Target path is not a directory")
		os.Exit(1)
	}

	buildFilePath := path.Join(CLI.Path, "PKGBUILD")
	buildFileContent, err := os.ReadFile(buildFilePath)
	if err != nil {
		i.Errorf("Failed to read build file: %s", err.Error())
		os.Exit(1)
	}

	var pkgbuild Pkgbuild
	_, err = toml.Decode(string(buildFileContent), &pkgbuild)
	if err != nil {
		i.Errorf("Failed to parse build file: %s", err.Error())
		os.Exit(1)
	}

	tempDirPath, err := os.MkdirTemp(os.TempDir(), "paw-checksums-*")
	if err != nil {
		i.Errorf("Failed to create temp dir: %s", err.Error())
		os.Exit(1)
	}

	baseDir, _ := filepath.Abs(CLI.Path)
	cwd, _ := os.Getwd()
	err = os.Chdir(tempDirPath)
	if err != nil {
		i.Errorf("Failed to change to temp dir: %s", err.Error())
		os.Exit(1)
	}

	i.Logf("Calculating checksums...")

	checksums := make(map[string]string)
	for _, src := range strings.Split(pkgbuild.Source, "\n") {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}

		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			filename := filepath.Base(src)
			i.Log2f("Downloading %s", filename)

			cmd := exec.Command("wget", "-q", src)
			cmd.Stderr = os.Stderr
			//cmd.Stdout = os.Stdout
			if err = cmd.Run(); err != nil {
				i.Error2f("Failed to download %s: %s", filename, err)
				os.Exit(1)
			}

			p := path.Join(tempDirPath, filename)
			cs, err := i.CalculateFileSHA256(p)
			if err != nil {
				i.Error2f("Failed to calculate checksum: %s", err.Error())
			}
			checksums[filename] = cs
			continue
		}

		srcPath := path.Join(baseDir, src)
		srcStat, err := os.Stat(srcPath)
		if err != nil {
			i.Error2f("Failed to stat %s: %s", srcPath, err)
			os.Exit(1)
		}

		if srcStat.IsDir() {
			_ = filepath.WalkDir(srcPath, func(_path string, d os.DirEntry, err error) error {
				if err != nil {
					i.Error2f("Failed to walk %s: %s", _path, err)
					os.Exit(1)
				}

				if d.IsDir() {
					return nil
				}

				relativeFilePath := strings.TrimPrefix(_path, baseDir+"/")
				cs, err := i.CalculateFileSHA256(_path)
				if err != nil {
					i.Error2f("Failed to calculate checksum: %s", err.Error())
				}
				checksums[relativeFilePath] = cs
				return nil
			})
		} else {
			cs, err := i.CalculateFileSHA256(srcPath)
			if err != nil {
				i.Error2f("Failed to calculate checksum: %s", err.Error())
			}
			checksums[src] = cs
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		i.Errorf("Failed to change directory: %s\n", err)
		os.Exit(1)
	}

	checksumFileContent := ""
	for filename, cs := range checksums {
		checksumFileContent += cs + " " + filename + "\n"
	}

	err = os.WriteFile(path.Join(CLI.Path, "checksums.txt"), []byte(checksumFileContent), 0644)
	if err != nil {
		i.Errorf("Failed to write checksums.txt: %s\n", err)
		os.Exit(1)
	}

	err = os.RemoveAll(tempDirPath)
	if err != nil {
		i.Errorf("Failed to remove temporary directory: %s\n", err)
		os.Exit(1)
	}
}
