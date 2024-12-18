package main

import (
	"os"

	"github.com/ocp-power-demos/power-dev-plugin/pkg/plugin"
	"k8s.io/klog"
)

// Launch the Plugin
func main() {
	devicePlugin, err := plugin.New()
	if err != nil {
		klog.V(2).Infof("Could not create new plugin, aborting")
		os.Exit(2)
	}
	if err := devicePlugin.Serve(); err != nil {
		klog.V(2).Infof("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
		os.Exit(3)
	}
	plugin.SystemShutdown()
}
