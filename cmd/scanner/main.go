package main

import (
	"os"

	"github.com/ocp-power-demos/power-dev-plugin/pkg/plugin"
	"k8s.io/klog"
)

// Launch the scanner
func main() {
	devices, err := plugin.ScanRootForDevices()
	if err != nil {
		klog.Errorf("Could not scan devices, aborting %s", err)
		os.Exit(2)
	}
	for idx, device := range devices {
		klog.Infof("%d - %s", idx, device)
	}
}
