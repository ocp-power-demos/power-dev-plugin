package main

import (
	"os"

	"github.com/ocp-power-demos/power-dev-plugin/pkg/webhook"
	"k8s.io/component-base/cli"
)

func main() {
	command := webhook.NewWebhookCommand()
	code := cli.Run(command)
	os.Exit(code)
}
