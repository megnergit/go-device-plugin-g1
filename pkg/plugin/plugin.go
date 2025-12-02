package plugin

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	dp "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resourceName = "example.com/mydevice"
	socketName   = "example.sock"
)

type DevicePlugin struct {
	dp.UnimplementedDevicePluginServer
	server *grpc.Server
}

func NewDevicePlugin() *DevicePlugin {
	return &DevicePlugin{}
}

func (p *DevicePlugin) Start() error {

	socketPath := filepath.Join(dp.DevicePluginPath, socketName)

	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	l, err := net.Listen("unix", socketPath)

	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	p.server = grpc.NewServer()
	dp.RegisterDevicePluginServer(p.server, p)

	go func() {
		log.Printf("Starting gRPC server on %s", socketPath)
		if err := p.server.Serve(l); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	conn, err := grpc.Dial(
		dp.KubeletSocket,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context,
			addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to kubelet: %w", err)
	}

	defer conn.Close()

	client := dp.NewRegistrationClient(conn)

	req := &dp.RegisterRequest{
		Version:      dp.Version,
		Endpoint:     socketName,
		ResourceName: resourceName,
	}

	if _, err := client.Register(context.Background(), req); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	log.Println("Device Plugin registered successfully.")
	return nil
}

func (p *DevicePlugin) ListAndWatch(_ *dp.Empty,
	stream dp.DevicePlugin_ListAndWatchServer) error {
	devices := []*dp.Device{
		{ID: "dev1", Health: dp.Healthy},
	}
	return stream.Send(&dp.ListAndWatchResponse{Devices: devices})
}

func (p *DevicePlugin) GetPreferredAllocation(
	ctx context.Context,
	reqs *dp.PreferredAllocationRequest,
) (*dp.PreferredAllocationResponse, error) {

	resp := &dp.PreferredAllocationResponse{
		ContainerResponses: []*dp.ContainerPreferredAllocationResponse{
			{
				DeviceIDs: []string{"dev1"},
			},
		},
	}
	return resp, nil
}

func (p *DevicePlugin) Allocate(
	ctx context.Context,
	reqs *dp.AllocateRequest,
) (*dp.AllocateResponse, error) {

	resp := &dp.AllocateResponse{}

	for range reqs.ContainerRequests {
		resp.ContainerResponses = append(resp.ContainerResponses,
			&dp.ContainerAllocateResponse{
				Envs: map[string]string{
					"MYDEVICE_ENABLED": "1",
				},
				Devices: []*dp.DeviceSpec{
					{
						HostPath:      "/dev/null",
						ContainerPath: "/dev/null",
						Permissions:   "rw",
					},
				},
			})
	}

	return resp, nil
}

func (p *DevicePlugin) GetDevicePluginOptions(context.Context,
	*dp.Empty) (*dp.DevicePluginOptions, error) {
	return &dp.DevicePluginOptions{}, nil
}

func (p *DevicePlugin) PreStartContainer(

	ctx context.Context,
	req *dp.PreStartContainerRequest,
) (*dp.PreStartContainerResponse, error) {

	return &dp.PreStartContainerResponse{}, nil
}

//----
// END
//----
