package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit"
)

const APT = "apt"

func main() {
	packit.Detect(detectFunc())
}

func detectFunc() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := os.Stat(filepath.Join(context.WorkingDir, "apt.yml"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to stat apt.yml: %w", err)
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: APT},
				},
				Requires: []packit.BuildPlanRequirement{
					{Name: APT},
				},
			},
		}, nil
	}
}
