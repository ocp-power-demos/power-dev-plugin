package plugin

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jaypipes/ghw"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resourceName = "ibm.com/power-dev"
	// SockDir is the default Kubelet device plugin socket directory
	SockDir    = pluginapi.DevicePluginPath
	pSocket    = "power-dev.sock"
	serverSock = SockDir + pSocket
	unix       = "unix"
)

// DevicePluginServer is a mandatory interface that must be implemented by all plugins.
// For more information see
// https://godoc.org/k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta#DevicePluginServer
type PowerPlugin struct {
	devs   []string
	socket string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server

	pluginapi.DevicePluginServer
}

// Creates a Plugin
func New() (*PowerPlugin, error) {
	// Empty array to start.
	var devs []string = []string{}
	return &PowerPlugin{
		devs:   devs,
		socket: serverSock,
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}, nil
}

// no-action needed to get options
func (m *PowerPlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired:                false,
		GetPreferredAllocationAvailable: false,
	}, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial() (*grpc.ClientConn, error) {
	c, err := grpc.NewClient(
		unix+":"+pluginapi.KubeletSocket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		klog.Errorf("%s device plugin unable connect to Kubelet : %v", serverSock, err)
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *PowerPlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		klog.Errorf("failed to listen on socket: %s", err.Error())
		return err
	}

	m.server = grpc.NewServer()
	pluginapi.RegisterDevicePluginServer(m.server, m)

	// start serving from grpcServer
	go func() {
		err := m.server.Serve(sock)
		if err != nil {
			klog.Errorf("serving incoming requests failed: %s", err.Error())
		}
	}()

	// Wait for server to start by launching a blocking connection
	// conn, err := dial(m.socket, 10*time.Second)
	// if err != nil {
	// 	klog.Errorf("unable to dial %v", err)
	// 	return err
	// }
	// conn.Close()

	// go m.healthcheck()

	return nil
}

// Stop stops the gRPC server
func (m *PowerPlugin) Stop() error {
	if m.server == nil {
		return nil
	}
	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// Registers the device plugin for the given resourceName with Kubelet.
func (m *PowerPlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial()
	defer conn.Close()
	if err != nil {
		return err
	}
	klog.Infof("Dial kubelet endpoint %s", conn.Target())

	client := pluginapi.NewRegistrationClient(conn)
	request := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pSocket,
		ResourceName: fmt.Sprintf("%s/%s", "power-dev-plugin", "dev"),
	}

	_, err = client.Register(context.Background(), request)
	if err != nil {
		return err
	}

	return nil
}

// Lists devices and update that list according to the health status
func (m *PowerPlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	klog.Infof("Exposing devices: %v", m.devs)

	if len(m.devs) == 0 {
		devices, err := ScanRootForDevices()
		if err != nil {
			klog.Errorf("Scan root for devices was unsuccessful during ListAndWatch: %v", err)
			return err
		} else {
			m.devs = devices
		}
	}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: convertDeviceToPluginDevices(m.devs)})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			// FIXME: there is no way to recover from the Unhealthy state.
			d.Health = pluginapi.Unhealthy
			s.Send(&pluginapi.ListAndWatchResponse{Devices: convertDeviceToPluginDevices(m.devs)})
		}
	}
}

// Allocate which return list of devices.
func (m *PowerPlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	klog.Infof("Allocate request: %v", reqs)

	devices, err := ScanRootForDevices()
	if err != nil {
		klog.Errorf("Scan root for devices was unsuccessful: %v", err)
		return nil, err
	}
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		klog.Infoln("Container requests device: ", req)
		ds := make([]*pluginapi.DeviceSpec, len(devices))
		response := pluginapi.ContainerAllocateResponse{Devices: ds}

		// Originally req.DeviceIds
		for i := range devices {
			ds[i] = &pluginapi.DeviceSpec{
				HostPath:      devices[i],
				ContainerPath: devices[i],
				// Per DeviceSpec:
				// Cgroups permissions of the device, candidates are one or more of
				// * r - allows container to read from the specified device.
				// * w - allows container to write to the specified device.
				// * m - allows container to create device files that do not yet exist.
				// We don't need `m`
				Permissions: "rw",
			}
		}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}
	klog.Infof("Allocate response: %v", responses)
	return &responses, nil
}

func convertDeviceToPluginDevices(devS []string) []*pluginapi.Device {
	klog.Infoln("Converting Devices to Plugin Devices")
	devs := []*pluginapi.Device{}
	for idx := range devS {
		devs = append(devs, &pluginapi.Device{
			ID:     strconv.Itoa(idx),
			Health: "healthy",
		})
	}
	klog.Infoln("Conversion completed")
	return devs
}

func (m *PowerPlugin) unhealthy(dev *pluginapi.Device) {
	m.health <- dev
}

// no-action needed to configure/load et cetra
func (m *PowerPlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (m *PowerPlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *PowerPlugin) Serve() error {
	err := m.Start()
	if err != nil {
		klog.Errorf("Could not start device plugin: %v", err)
		return err
	}
	klog.Infof("Starting to serve on %s", m.socket)

	err = m.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		klog.Errorf("Could not register device plugin: %v", err)
		m.Stop()
		return err
	}
	klog.Infof("Registered device plugin with Kubelet")
	return nil
}

// func (p *PowerPlugin) Run() int {
// 	klog.V(0).Info("Start Run")
// 	stopCh := make(chan struct{})
// 	defer close(stopCh)

// 	exitCh := make(chan error)
// 	defer close(exitCh)

// 	for {
// 		select {
// 		case <-stopCh:
// 			klog.V(0).Info("Run(): stopping plugin")
// 			return 0
// 		case err := <-exitCh:
// 			klog.Error(err, "got an error", err)
// 			return 99
// 		}
// 	}
// }

// Kublet may restart, and we'll need to restart.
// func monitorPluginRegistration() error {
// 	return nil
// }

// Captures the Signal to shutdown the container and dispatches to the Application
func SystemShutdown() {
	// Get notified about syscall
	klog.V(1).Infof("Listening for term signals")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Catch termination signals
	sig := <-sigCh
	klog.Infof("Received signal \"%v\", shutting down.", sig)
	if err := AppShutdown(); err != nil {
		klog.Errorf("stopping servers produced error: %s", err.Error())
	}
}

// Shutdown the Application
func AppShutdown() error {
	return nil
}

// scans the local disk using ghw to find the blockdevices
func ScanRootForDevices() ([]string, error) {
	// relies on GHW_CHROOT=/host/dev
	// lsblk -f --json --paths -s | jq -r '.blockdevices[] | select(.fstype != "xfs")' | grep mpath | grep -v fstype | sort -u | wc -l
	// This may be the best way to get the devices.
	block, err := ghw.Block()
	if err != nil {
		klog.Errorf("Error getting block storage info: %v", err)
		return nil, err
	}
	devices := []string{}
	klog.Infof("%v\n", block)
	for _, disk := range block.Disks {
		klog.Infof(" %v\n", disk)
		for _, part := range disk.Partitions {
			klog.Infof("  %v\n", part)
			devices = append(devices, part.Name)
		}
	}
	return devices, nil
}
