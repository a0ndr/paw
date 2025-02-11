package main

import (
	"github.com/BurntSushi/toml"
	_p "github.com/a0ndr/paw/pkg"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	cp "github.com/otiai10/copy"
	"lure.sh/fakeroot"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var CLI struct {
	Config string `flag:"config" short:"C" help:"Path to config file" default:"/etc/paw.toml"`

	NoCleanup bool `flag:"" help:"Do not clean up"`

	Path string `arg:"" help:"Package to build" default:"."`
}

func _logf(format string, args ...interface{}) {
	_, _ = color.New(color.FgYellow).Add(color.Bold).Print("==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}

func __logf(format string, args ...interface{}) {
	_, _ = color.New(color.FgCyan).Add(color.Bold).Print("  ==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}

func _errorf(format string, args ...interface{}) {
	_, _ = color.New(color.FgRed).Add(color.Bold).Print("==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}

func __errorf(format string, args ...interface{}) {
	_, _ = color.New(color.FgRed).Add(color.Bold).Print("  ==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
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
	_p.LoadConfig(CLI.Config)

	if stat, err := os.Stat(CLI.Path); err != nil || !stat.IsDir() {
		_errorf("The target package path is not a directory")
		os.Exit(1)
	}

	buildFilePath := path.Join(CLI.Path, "PKGBUILD")
	buildFileContent, err := os.ReadFile(buildFilePath)
	if err != nil {
		_errorf("Failed to read PKGBUILD: %s", err)
		os.Exit(1)
	}

	var pkgbuild Pkgbuild
	_, err = toml.Decode(string(buildFileContent), &pkgbuild)
	if err != nil {
		_errorf("Failed to parse PKGBUILD: %s", err)
		os.Exit(1)
	}

	_logf("Building package %s-%s", pkgbuild.Name, pkgbuild.Version)
	_logf("Creating build directories")

	tempDirPath, err := os.MkdirTemp(os.TempDir(), "paw-build-*")
	if err != nil {
		__errorf("Failed to create temporary directory: %s", err)
		os.Exit(1)
	}

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
		__errorf("Failed to create directory: %s", err)
		os.Exit(1)
	}

	// GATHER STUFF
	_logf("Gathering sources")
	err = os.Chdir(srcDir)
	if err != nil {
		__errorf("Failed to change directory: %s", err)
		os.Exit(1)
	}

	for _, src := range strings.Split(pkgbuild.Source, "\n") {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}

		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			__logf("Downloading %s...", src)

			cmd := exec.Command("wget", src)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err = cmd.Run(); err != nil {
				__errorf("Failed to download %s: %s", src, err)
				os.Exit(1)
			}

			continue
		}

		srcPath := path.Join(baseDir, src)
		_, err := os.Stat(srcPath)
		if err != nil {
			__errorf("Failed to stat %s: %s", srcPath, err)
			os.Exit(1)
		}

		err = cp.Copy(srcPath, path.Join(srcDir, filepath.Base(srcPath)))
		if err != nil {
			__errorf("Failed to copy %s: %s", srcPath, err)
			os.Exit(1)
		}
	}

	_logf("Building package")

	buildScriptPath := path.Join(tempDirPath, "build.sh")
	err = os.WriteFile(buildScriptPath, []byte("#!/usr/bin/env sh\n"+pkgbuild.Build), 0744)
	if err != nil {
		__errorf("Failed to write build script: %s", err)
		os.Exit(1)
	}

	__logf("Creating fakeroot")
	cmd, err := fakeroot.Command(buildScriptPath)
	if err != nil {
		__errorf("Failed to make fakeroot: %s", err)
		os.Exit(1)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "SRCDIR="+srcDir)
	cmd.Env = append(cmd.Env, "OUTDIR="+outDir)
	cmd.Env = append(cmd.Env, "PKGDIR="+pkgDir)

	__logf("Executing build script")
	err = cmd.Run()
	if err != nil {
		__errorf("Failed to build package: %s", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}

	if CLI.NoCleanup {
		return
	}

	// CLEANUP
	_logf("Cleaning up")
	err = os.RemoveAll(tempDirPath)
	if err != nil {
		__errorf("Failed to remove temporary directory: %s", err)
		os.Exit(1)
	}
}
