package plugin

import (
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Message struct {
	syncStatus    string
	lastSyncError string
}

type Plugin struct{}

// Creates a Plugin
func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Run() int {
	log.Log.V(0).Info("Start Ru")
	stopCh := make(chan struct{})
	defer close(stopCh)

	exitCh := make(chan error)
	defer close(exitCh)

	for {
		select {
		case <-stopCh:
			log.Log.V(0).Info("Run(): stopping plugin")
			return 0
		case err := <-exitCh:
			log.Log.Error(err, "got an error", err)
			return 99
		}
	}
}

// Kublet may restart, and we'll need to restart.
func monitorPluginRegistration() error {
	return nil
}
