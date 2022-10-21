//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/sh"
	"hermannm.dev/bfh-server/magefiles/color"
)

func CrossCompile() error {
	appName := "bfh-server"
	packages := []string{"local", "public"}
	packageDir := "./cmd"
	outputDir := "bin"

	platforms := map[string][]string{
		"darwin":  {"amd64", "arm64"},
		"linux":   {"386", "amd64", "arm64"},
		"windows": {"386", "amd64", "arm64"},
	}

	for _, pkg := range packages {
		inputLocation := fmt.Sprintf("%s/%s", packageDir, pkg)

		for os, architectures := range platforms {
			for _, arch := range architectures {
				binName := fmt.Sprintf("%s_%s_%s-%s", appName, pkg, os, arch)
				if os == "windows" {
					binName += ".exe"
				}

				outputLocation := fmt.Sprintf("%s/%s/%s", outputDir, pkg, binName)

				env := map[string]string{"GOOS": os, "GOARCH": arch}

				fmt.Printf("%s %s\n", color.Blue.String("[Building]"), binName)

				if err := sh.RunWithV(env, "go", "build", "-o", outputLocation, inputLocation); err != nil {
					return fmt.Errorf("cross-compilation failed: %w", err)
				}
			}
		}
	}

	fmt.Printf("%s Output in: %s\n", color.Green.String("[Finished]"), outputDir)

	return nil
}
