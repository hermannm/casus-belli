//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/sh"
	"hermannm.dev/bfh-server/magefiles/color"
	"hermannm.dev/wrap"
)

func CrossCompile() error {
	appName := "bfh-server"
	outputDir := "bin"

	platforms := map[string][]string{
		"darwin":  {"amd64", "arm64"},
		"linux":   {"386", "amd64", "arm64"},
		"windows": {"386", "amd64", "arm64"},
	}

	for os, architectures := range platforms {
		for _, arch := range architectures {
			binName := fmt.Sprintf("%s-%s-%s", appName, os, arch)
			if os == "windows" {
				binName += ".exe"
			}

			outputLocation := fmt.Sprintf("%s/%s", outputDir, binName)

			env := map[string]string{"GOOS": os, "GOARCH": arch}

			fmt.Printf("%s %s\n", color.Blue.String("[Building]"), binName)

			if err := sh.RunWithV(env, "go", "build", "-o", outputLocation); err != nil {
				return wrap.Error(err, "cross-compilation failed")
			}
		}
	}

	fmt.Printf("%s Output in: %s\n", color.Green.String("[Finished]"), outputDir)

	return nil
}
