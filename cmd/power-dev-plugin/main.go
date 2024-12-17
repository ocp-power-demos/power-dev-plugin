package main

import (
	"os"

	"github.com/ocp-power-demos/power-dev-plugin/pkg/plugin"
)

func main() {
	os.Exit(plugin.New().Run())
}
