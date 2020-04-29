// +build mage

package main

import (
	"bufio"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/yinheli/database-struct/version"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

var app = version.AppName

// tidy code
func Fmt() error {
	packages := strings.Split("cmd pkg version", " ")
	return sh.Run("gofmt", append([]string{"-s", "-l", "-w", "mage.go", "magefile.go"}, packages...)...)
}

// for local machine build
func Build() error {
	return buildTarget(runtime.GOOS, runtime.GOARCH, nil)
}

// for local linux
func Linux() error {
	return buildTarget("linux", "amd64", nil)
}

// for macos
func MacOs() error {
	return buildTarget("darwin", "amd64", nil)
}

// for windows
func Windows() error {
	return buildTarget("windows", "amd64", nil)
}

// golangci-lint
func Lint() error {
	return sh.RunV("golangci-lint", "run")
}

// build to target (cross build)
func buildTarget(OS, arch string, envs map[string]string) error {
	name := version.AppName
	dir := fmt.Sprintf("dist/%s-%s-%s", name, OS, arch)
	target := fmt.Sprintf("%s/%s", dir, app)

	args := make([]string, 0, 10)
	args = append(args, "build", "-o", target)
	args = append(args, "-ldflags="+flags(), "cmd/main.go")

	fmt.Println("build", name)
	env := make(map[string]string)
	env["GOOS"] = OS
	env["GOARCH"] = arch

	if envs != nil {
		for k, v := range envs {
			env[k] = v
		}
	}

	if err := sh.RunWith(env, mg.GoCmd(), args...); err != nil {
		return err
	}

	_ = sh.Run("tar", "-czf", fmt.Sprintf("%s.tar.gz", dir), "-C", "dist", filepath.Base(dir))
	return nil
}

func flags() string {
	timestamp := time.Now().Format(time.RFC3339)
	h := hash()
	m := mod()
	tpl := fmt.Sprintf(`-s -w -X "%s/version.Build=%%s" -X "%s/version.BuildAt=%%s"`, m, m)
	build := fmt.Sprintf("%s", h)
	return fmt.Sprintf(tpl, build, timestamp)
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	if hash == "" {
		return "0000000"
	}
	return hash
}

func mod() string {
	f, err := os.Open("go.mod")
	if err == nil {
		reader := bufio.NewReader(f)
		line, _, _ := reader.ReadLine()
		return strings.Replace(string(line), "module ", "", 1)
	}
	return ""
}

// cleanup all build files
func Clean() {
	_ = sh.Rm("dist")
}
