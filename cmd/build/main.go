package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/packit"
	"gopkg.in/yaml.v2"
)

type aptYAML struct {
	Packages []string `yaml:packages`
}

const APT = "apt"

func main() {
	packit.Build(buildFunc())
}

func buildFunc() packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		aptLayer, err := context.Layers.Get(APT, packit.LaunchLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		aptConfig := aptYAML{}
		configFile, err := os.Open(filepath.Join(context.WorkingDir, "apt.yml"))
		if err != nil {
			return packit.BuildResult{}, err
		}
		if err := yaml.NewDecoder(configFile).Decode(&aptConfig); err != nil {
			return packit.BuildResult{}, err
		}

		if err := aptLayer.Reset(); err != nil {
			return packit.BuildResult{}, err
		}

		/**
		 * ref: https://github.com/cloudfoundry/apt-buildpack/blob/master/src/apt/apt/apt.go
		 */
		cacheDir := "/tmp"
		aptCacheDir := filepath.Join(cacheDir, "apt", "cache")
		stateDir := filepath.Join(cacheDir, "apt", "state")

		archiveDir := filepath.Join(aptCacheDir, "archives")
		installDir := aptLayer.Path

		globalOptions := []string{
			"-o", "debug::nolocking=true",
			"-o", "dir::cache=" + aptCacheDir,
			"-o", "dir::state=" + stateDir,
		}

		if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
			return packit.BuildResult{}, err
		}

		if err := os.MkdirAll(archiveDir, os.ModePerm); err != nil {
			return packit.BuildResult{}, err
		}

		if err := os.MkdirAll(stateDir, os.ModePerm); err != nil {
			return packit.BuildResult{}, err
		}

		if err := executeCommand("apt-get", append(globalOptions, []string{"update", "-y"}...)); err != nil {
			return packit.BuildResult{}, err
		}

		if err := executeCommand("apt-get", append(append(globalOptions, []string{"install", "-y", "-d"}...), aptConfig.Packages...)); err != nil {
			return packit.BuildResult{}, err
		}

		files, err := filepath.Glob(filepath.Join(archiveDir, "*.deb"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		for _, file := range files {
			pkg := filepath.Join(archiveDir, filepath.Base(file))
			if err := executeCommand("dpkg", []string{"-x", pkg, installDir}); err != nil {
				return packit.BuildResult{}, err
			}
		}

		return packit.BuildResult{
			Plan: context.Plan,
			Layers: []packit.Layer{
				aptLayer,
			},
		}, nil
	}
}

func executeCommand(c string, opts []string) error {
	cmd := exec.Command(c, opts...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
