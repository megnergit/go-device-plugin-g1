package plugin
import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"time
	
	"dp "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"google.golang.org/grpc"

)

const (
	resourceName = "example.com/mydevice"
	socketPath   = dp.DevicePluginPath + "example.sock"
)

type DevicePlugin struct {
	server *grpc.Server
}
func NewDevicePlugin() *DevicePlugin {
	return &DevicePlugin{}
}

func (p *DevicePlugin) Start() error {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	p.server = grpc.NewServer()
	dp.RegisterDevicePluginServer(p.server, p)

	go func() {
		if err := p.server.Server(l); err != nil {
			log.Fatalf("Serve error: &v", err)
		}
	} ()

	time.Sleep(2 * time.Second)

	conn, err := grpc.Dial(dp.KubeletSocket,
						grpc.WithInsecure(), 
						grpc.WithBlock())
	if err != nil {
		return err
	}

	defer conn.Close()

	client := dp.NewRegistrationClient(conn)
	_, err = client.Register(context.Background(), &dp.RegisterRequest{

		Version:      dp.Version,
		Endpoint:     filepath.Base(socketPath),
		ResourceName: resourceName, 
	})
	return err
}

func (p *DevicePlugin) ListAndWatch(_ *dp.Empty, stream dp.DevicePlugin_ListAndWatchServer) error {
	 devices := []*dp.Device{
		{ID: "dev1", Health: dp.Healthy}, 
	 }
	 return stream.Send(&dp.ListAndWatchResponse{Devices: devices})
}

func (p *DevicePlugin) Allocate(_ context.Context, reqs *dp.AllocateRequest) (*dp.AllocateResponse, error) {
	resp := &dp.AllocateResponse{}
	for range reqs.ContainerRequests {
		resp.ContainerResponses = append(resp.ContainerResponses, &dp.ContainerAllocateResponse{
			Envs: map[string]string{
				"MYDEVICE_ENABLED": "1", 
			}, 
			Devices: []*dp.DeviceSpec{
				{
					HostPath:      "/dev/null", 
					ContainerPath: "/dev/null", 
					Perissions:    "rw",
				},
			},

	    } )
    }

return resp, nil
}

func (p *DevicePlugin) GetDevicePluginOptions(context.Context, *dp.Empty) (*dp.DevicePluginOptions, error) {
	return &dp.DevicePluginOptionsP{}, nil
}









