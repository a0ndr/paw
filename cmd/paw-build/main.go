package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/BurntSushi/toml"
	i "github.com/a0ndr/paw/internal"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	cp "github.com/otiai10/copy"
	"io"
	"lure.sh/fakeroot"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`
	Output string `flag:"" short:"o" help:"Path to output directory" default:"."`

	NoCleanup bool `flag:"" help:"Do not clean up"`

	Path string `arg:"" help:"Package to build" default:"."`
}

var tempDirPath string

func cleanup(code int) {
	if !CLI.NoCleanup {
		i.Logf("Cleaning up...")
		err := os.RemoveAll(tempDirPath)
		if err != nil {
			i.Error2f("Failed to remove temporary directory: %s", err)
		}
	}
	os.Exit(code)
}

func createTarball(srcDir, output string) error {
	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	gzWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer func(gzWriter *gzip.Writer) {
		_ = gzWriter.Close()
	}(gzWriter)

	tarWriter := tar.NewWriter(gzWriter)
	defer func(tarWriter *tar.Writer) {
		_ = tarWriter.Close()
	}(tarWriter)

	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if file == srcDir {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		srcFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer func(srcFile *os.File) {
			_ = srcFile.Close()
		}(srcFile)

		_, err = io.Copy(tarWriter, srcFile)
		return err
	})
}

func main() {
	_ = kong.Parse(&CLI)
	_p.LoadConfig(CLI.Config)

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
		i.Errorf("Failed to read PKGBUILD: %s", err)
		os.Exit(1)
	}

	var pkgbuild _p.Build
	_, err = toml.Decode(string(buildFileContent), &pkgbuild)
	if err != nil {
		i.Errorf("Failed to parse PKGBUILD: %s", err)
		os.Exit(1)
	}

	i.Logf("Building package %s-%s...", pkgbuild.Name, pkgbuild.Version)
	i.Logf("Creating build directories")

	tempDirPath, err = os.MkdirTemp(os.TempDir(), "paw-build-*")
	if err != nil {
		i.Error2f("Failed to create temporary directory: %s", err)
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	baseDir, _ := filepath.Abs(CLI.Path)
	srcDir := path.Join(tempDirPath, "src") // downloading stuff, deps
	//buildDir := path.Join(tempDirPath, "build") // compiling
	outDir := path.Join(tempDirPath, "out") // compile output
	pkgDir := path.Join(tempDirPath, "pkg") // final package directory, this will be compressed

	err = os.Mkdir(srcDir, 0755)
	//err = os.Mkdir(buildDir, 0755)
	err = os.Mkdir(outDir, 0755)
	err = os.Mkdir(pkgDir, 0755)

	if err != nil {
		i.Error2f("Failed to create directory: %s", err)
		cleanup(1)
	}

	// GATHER STUFF
	i.Logf("Gathering sources...")
	err = os.Chdir(srcDir)
	if err != nil {
		i.Error2f("Failed to change directory: %s", err)
		cleanup(1)
	}

	for _, src := range strings.Split(pkgbuild.Source, "\n") {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}

		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			i.Log2f("Downloading %s", src)

			cmd := exec.Command("wget", src)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err = cmd.Run(); err != nil {
				i.Error2f("Failed to download %s: %s", src, err)
				cleanup(1)
			}

			continue
		}

		srcPath := path.Join(baseDir, src)
		_, err = os.Stat(srcPath)
		if err != nil {
			i.Error2f("Failed to stat %s: %s", srcPath, err)
			cleanup(1)
		}

		err = cp.Copy(srcPath, path.Join(srcDir, filepath.Base(srcPath)))
		if err != nil {
			i.Error2f("Failed to copy %s: %s", srcPath, err)
			cleanup(1)
		}
	}

	i.Logf("Validating checksums...")

	checksums := make(map[string]string)
	checksumsFile, err := os.ReadFile(path.Join(baseDir, "checksums.txt"))
	if err != nil {
		i.Error2f("Failed to read checksums.txt: %s", err)
		i.Log2f("Do you need to run paw-checksum?")
		cleanup(1)
	}

	for _, line := range strings.Split(string(checksumsFile), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		split := strings.SplitN(line, " ", 2)
		file := split[1]
		checksum := split[0]

		absFilePath := path.Join(srcDir, file)
		cs, err := i.CalculateFileSHA256(absFilePath)
		if err != nil {
			i.Error2f("Failed to calculate checksum: %s", err)
			cleanup(1)
		}

		if cs != checksum {
			i.Error2f("Checksum for file \"%s\" does not match", file)
			cleanup(1)
		}

		checksums[file] = cs
	}

	i.Log2f("Checksums match!")
	i.Logf("Building package...")

	buildScriptPath := path.Join(tempDirPath, "BUILD.sh")
	err = os.WriteFile(buildScriptPath, []byte("#!/usr/bin/env sh\n"+pkgbuild.Build), 0744)
	if err != nil {
		i.Error2f("Failed to write build script: %s", err)
		cleanup(1)
	}

	i.Log2f("Creating fakeroot...")
	cmd, err := fakeroot.Command(buildScriptPath)
	if err != nil {
		i.Error2f("Failed to make fakeroot: %s", err)
		cleanup(1)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "SRCDIR="+srcDir)
	cmd.Env = append(cmd.Env, "OUTDIR="+outDir)
	cmd.Env = append(cmd.Env, "PKGDIR="+pkgDir)

	i.Log2f("Executing build script...")
	if err = cmd.Run(); err != nil {
		i.Error2f("Failed to build package: %s", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}

	// PACKAGE
	i.Logf("Packaging...")
	i.Log2f("Generating package metadata...")

	pkgMetaPath := path.Join(pkgDir, "PKGMETA")
	f, err := os.OpenFile(pkgMetaPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		i.Error2f("Fatal: could not open file for writing: %v\n", err)
		cleanup(1)
	}

	pkgMeta := &_p.Meta{
		Name:             pkgbuild.Name,
		Version:          pkgbuild.Version,
		Description:      pkgbuild.Description,
		Checksums:        checksums,
		Dependencies:     pkgbuild.Dependencies,
		SoftDependencies: pkgbuild.SoftDependencies,
		Conflicts:        pkgbuild.Conflicts,
		BuiltAt:          time.Now(),
	}

	decoder := toml.NewEncoder(f)
	err = decoder.Encode(pkgMeta)
	if err != nil {
		i.Error2f("Failed to encode PKGMETA: %s", err)
		cleanup(1)
	}
	_ = f.Close()

	i.Log2f("Writing scripts...")

	if pkgbuild.PreInstall != "" {
		err = os.WriteFile(path.Join(pkgDir, "PREINSTALL.sh"), []byte(fmt.Sprintf("#!/usr/bin/env sh\n%s", pkgbuild.PreInstall)), 0644)
		if err != nil {
			i.Error2f("Failed to write pre-install script: %s", err)
			cleanup(1)
		}
	}

	err = os.WriteFile(path.Join(pkgDir, "INSTALL.sh"), []byte(fmt.Sprintf("#!/usr/bin/env sh\n%s", pkgbuild.Install)), 0644)
	if err != nil {
		i.Error2f("Failed to write install script: %s", err)
		cleanup(1)
	}

	if pkgbuild.PostInstall != "" {
		err = os.WriteFile(path.Join(pkgDir, "POSTINSTALL.sh"), []byte(fmt.Sprintf("#!/usr/bin/env sh\n%s", pkgbuild.PostInstall)), 0644)
		if err != nil {
			i.Error2f("Failed to write post-install script: %s", err)
			cleanup(1)
		}
	}

	err = os.WriteFile(path.Join(pkgDir, "CONFIGURE.sh"), []byte(fmt.Sprintf("#!/usr/bin/env sh\n%s", pkgbuild.Configure)), 0644)
	if err != nil {
		i.Error2f("Failed to write configure script: %s", err)
		cleanup(1)
	}

	if pkgbuild.PreRemove != "" {
		err = os.WriteFile(path.Join(pkgDir, "PREREMOVE.sh"), []byte(fmt.Sprintf("#!/usr/bin/env sh\n%s", pkgbuild.PreRemove)), 0644)
		if err != nil {
			i.Error2f("Failed to write pre-remove script: %s", err)
			cleanup(1)
		}
	}

	if pkgbuild.Remove != "" {
		err = os.WriteFile(path.Join(pkgDir, "REMOVE.sh"), []byte(fmt.Sprintf("#!/usr/bin/env sh\n%s", pkgbuild.Remove)), 0644)
		if err != nil {
			i.Error2f("Failed to write remove script: %s", err)
			cleanup(1)
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		i.Error2f("Failed to change directory: %s", err)
		cleanup(1)
	}

	i.Log2f("Creating tarball...")

	tarballPath := path.Join(CLI.Output, fmt.Sprintf("%s-%s.tar.gz", pkgbuild.Name, pkgbuild.Version))
	err = createTarball(pkgDir, tarballPath)
	if err != nil {
		i.Error2f("Failed to create tarball: %s", err)
		cleanup(1)
	}

	if !CLI.NoCleanup {
		i.Logf("Cleaning up...")
		err := os.RemoveAll(tempDirPath)
		if err != nil {
			i.Error2f("Failed to remove temporary directory: %s", err)
			os.Exit(1)
		}
	}

	i.Logf("Package built as %s", tarballPath)
}
