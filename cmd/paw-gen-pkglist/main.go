package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/BurntSushi/toml"
	i "github.com/a0ndr/paw/internal"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var CLI struct {
	Output string `flag:"" short:"o" help:"Output directory path" default:"."`
	Path   string `arg:"" help:"Package directory path" default:"."`
}

func readMetaFromTarball(tarball string) (*_p.Meta, error) {
	file, err := os.Open(tarball)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer func(gzReader *gzip.Reader) {
		_ = gzReader.Close()
	}(gzReader)

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "PKGMETA" {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}

			var meta _p.Meta
			_, err = toml.Decode(string(data), &meta)
			if err != nil {
				return nil, err
			}
			return &meta, nil
		}
	}

	return nil, fmt.Errorf("package meta not found in archive")
}

func main() {
	_ = kong.Parse(&CLI)

	// sanity checks
	if stat, err := os.Stat(CLI.Path); err != nil || !stat.IsDir() {
		i.Errorf("%s is not a directory", CLI.Path)
		os.Exit(1)
	}
	if stat, err := os.Stat(CLI.Output); err != nil || !stat.IsDir() {
		i.Errorf("%s is not a directory", CLI.Output)
		os.Exit(1)
	}

	i.Logf("Generating package list...")
	packages := &_p.List{Packages: make(map[string]*_p.Entry)}

	contents, err := os.ReadDir(CLI.Path)
	if err != nil {
		i.Error2f("Could not read directory: %s", err.Error())
		os.Exit(1)
	}
	for _, file := range contents {
		if !strings.HasSuffix(file.Name(), ".tar.gz") {
			i.Log2f("Skipping %s", file.Name())
			continue
		}

		i.Log2f("Processing %s", file.Name())

		filePath := filepath.Join(CLI.Path, file.Name())
		meta, err := readMetaFromTarball(filePath)
		if err != nil {
			i.Error2f("Could not read package meta of package %s: %s", file.Name(), err.Error())
			os.Exit(1)
		}

		checksum, err := i.CalculateFileSHA256(filePath)
		if err != nil {
			i.Error2f("Could not calculate checksum of package %s: %s", file.Name(), err.Error())
		}

		packages.Packages[fmt.Sprintf("%s-%s", meta.Name, meta.Version)] = meta.ToEntry(checksum)
	}

	i.Logf("Saving package list...")

	outputPath := path.Join(CLI.Output, "PKGLIST")
	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		i.Error2f("Could not open package list: %s", err.Error())
		os.Exit(1)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	encoder := toml.NewEncoder(f)
	err = encoder.Encode(packages)
	if err != nil {
		i.Error2f("Could not encode package list: %s", err.Error())
		os.Exit(1)
	}
}
