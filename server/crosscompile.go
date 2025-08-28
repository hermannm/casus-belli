//go:build mage

// Uses Mage: https://magefile.org/
// After installing Mage, run it with "mage crosscompile"

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"golang.org/x/sync/errgroup"
	"hermannm.dev/devlog"
	"hermannm.dev/wrap"
)

var (
	appName   = "casus-belli-server"
	outputDir = "./bin"
	targets   = []struct {
		os            string
		architectures []string
	}{
		{os: "darwin", architectures: []string{"amd64", "arm64"}},
		{os: "linux", architectures: []string{"386", "amd64", "arm64"}},
		{os: "windows", architectures: []string{"386", "amd64", "arm64"}},
	}
)

func CrossCompile() error {
	var goroutines errgroup.Group

	for _, target := range targets {
		for _, arch := range target.architectures {
			targetString := target.os + "-" + arch
			fmt.Println(withColor("[Building]", blue), targetString)

			goroutines.Go(
				func() error {
					output := outputDir + "/" + appName + "-" + targetString
					if target.os == "windows" {
						output += ".exe"
					}
					env := map[string]string{"GOOS": target.os, "GOARCH": arch}
					return sh.RunWithV(env, "go", "build", "-o", output)
				},
			)
		}
	}

	if err := goroutines.Wait(); err != nil {
		return wrap.Error(err, "cross-compilation failed")
	}

	fmt.Println(withColor("[Finished]", green), "Output in:", outputDir)
	return nil
}

type color string

var (
	blue       color = "\x1b[34m"
	green      color = "\x1b[32m"
	resetColor color = "\x1b[0m"

	colorEnabled = devlog.IsColorTerminal(os.Stdout)
)

func withColor(str string, color color) string {
	if colorEnabled {
		return string(color) + str + string(resetColor)
	} else {
		return str
	}
}
