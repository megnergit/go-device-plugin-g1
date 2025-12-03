# Go Device Plugin 
Test device plugin with GoLang.

## Table of Contents

-[Motivation](#motivation)
- [Go Device Plugin](#go-device-plugin)
  - [Table of Contents](#table-of-contents)
  - [Motivation](#motivation)
    - [What is a device plugin?](#what-is-a-device-plugin)
    - [How a device plugin communicate with cluster](#how-a-device-plugin-communicate-with-cluster)
    - [Required methods in a device plugin](#required-methods-in-a-device-plugin)
    - [Goal](#goal)
  - [Techstack of choice](#techstack-of-choice)
    - [Kind](#kind)
    - [Container Runtime](#container-runtime)
    - [VM](#vm)
    - [Container runtime](#container-runtime-1)
  - [Build Infrastructure](#build-infrastructure)
    - [Kubernets](#kubernets)
  - [Directory Structure](#directory-structure)
  - [GoLang](#golang)
    - [Which version?](#which-version)
  - [Source Code Analysis](#source-code-analysis)
    - [```main.go```](#maingo)
    - [```plugin.go```](#plugingo)
  - [Test](#test)
- [END](#end)


---
## Motivation

This is the first experiment to understand how a device plugin 
works for Kubernetes. We will write one in GoLang. 

### What is a device plugin?

A device plugin for kubernetes is often explained that it bridges 
between a hardware of all kinds and a Kubernetes cluster. A device plugin 
makes the hardware visible to a pod. 

In fact 90% of the cases that hardware is a GPU, and the rest is IIoT 
devices or storage devices for which  official plugins are not yet available. 

A device driver does not drive a hardware by itself. That is what 
a **device driver** does, which is installed directly on the operation 
system of the baremetal. A device driver merely makes this hardware
visible to a pod running on a kubernetes cluster. 

We need to configure a device driver to operate a hardware. For instance 
in the case of GPU, we would need a device file (like ```/dev/nvidia0```), 
how many GPUs we will use, how we slice these GPUs, etc. A device plugin helps 
to configure these parameters from a kubernetes cluster. 

A device plugin runs as a **Daemonset**, which is a pod that runs all the nodes
that consists of a cluster. 

A device plugin is only required to run on the node where the hardware (most of 
the time, it is a GPU) is installed.  Therefore we need ```toleration``` or 
```nodeSelector``` so that the device plugin runs on the right node. 


### How a device plugin communicate with cluster

A device plugin works side by side with **kubelet**, and 
communicate only with kubelet. A device plugin does **not**
discuss with API server, scheduler, controller manager or a pod.

The comnunication protocol is **gRPC**, therefore a device driver 
is usually written in **GoLang**. 

```sh
+-----------------------------+
|         kubelet             |
|   /var/lib/kubelet/device-plugins/kubelet.sock
+------------┬----------------+
             │  gRPC
             │
+------------▼----------------+
|      Device Plugin          |
|  (your daemonset pod)       |
+-----------------------------+

```

A device plugin give kubelet access to the hardware

1. device plugin first let kubelet know the list of devices (ListAndWatch)
2. kubelet will advertise to scheduler that this node has that hardware
3. scheduler will assign a pod to that node, when that pod require that hardware
4. kubelet asks device plugin to assign the hardware to the pod (Allocate)
5. device plugins let kubelet know necessary infos such as env, device mount, etc.
6. kubelet uses these infos and create a container. 


### Required methods in a device plugin

What methods a device plugin should have is clearly defined by 
the Kubernetes [design](https://github.com/kubernetes/kubelet/blob/master/pkg/apis/deviceplugin/v1beta1/api.proto).


| Mandatory           |                                                          | 
|---------------------|--------------------------------------------------------| 
| ListAndWatch        | to show kubelet list of device and their health states | 
| Allocate            | to give kubelet env. variables and device mount infos. |
| GetDevicePluginOptions | to let kubelet know what the device can do           |    


| Optional (but almost mandatory)   |                                           |
|---------------------|--------------------------------------------------------| 
| GetPreferredAllocation | answer inquiry from kubelet, which device is optimal | 
| PreStartContainer      | prepare creating container                           |



### Goal

We will create a minimum device plugin, which 

-  runs on a kubernetes cluster as a **Daemonset**
-  advertises its presence to **kubelet** via **unix socket**
-  when called, hands **/dev/null** to a pod

---

## Techstack of choice

### Kind

As my laptop is recently overhauled (the system is reinstalled 
    from the operation system up), we have a complete free hand to select the tech-stuck. As a Kubernetes backbone, I would have following options. 

    - native kubernetes with kubeadm  
    - microk8s
    - minikube
    - kind
    - k3s

This time I opt to **kind**, as my laptop is now under performant (= hardware is tool old) for the tasks it is up to.


### Container Runtime

As docker desktop is too heavy for my machine, we will go for light-weight docker desktop alternative, [Colima](https://github.com/abiosoft/colima).

### VM

Therefore we will create backbone VMs with [Lima](https://github.com/lima-vm/lima).

### Container runtime

I prefer ```containerd``` to docker, but could not use ```nerdctl``` as it does not support macOS 13 ventura.
```sh

$ brew install nerdctl
Warning: You are using macOS 13.
We (and Apple) do not provide support for this old version.
You may have better luck with MacPorts which supports older versions of macOS:
  https://www.macports.org
...
```

Moreover, containerd requires qemu, and qemu  
dose not run on macOS 13 Ventura. Therefore I have to live with docker. Sigh... 


We will draw back to ```docker```, but at least without docker desktop.


---
## Build Infrastructure

We will build the architecture in the following order.

1. Lima
2. Colima
3. docker (without 'desktop') 
4. kind
5. GoLang


Roughly speaking,

|             |                   |
|-------------|-------------------|  
| docker      | docker            |
| Colima      | ~ docker desktop  |
| lima        | ~ vagrant         | 
| qemu        | ~ virtual box     |



So, like this. 

```sh
macOS
  └── Colima VM (Linux)
        └── Docker
              └── kind node (Kubernetes)
                    └── kubelet
                          └── your device plugin

```

Colima uses qemu inside, but we do not need to install qemu separately.

Okay, then first lima,
```sh
$ brew install lima
```

Check, 

```sh
$ lima --version
limactl version 2.0.2
```

Then docker. 

```sh
$ brew install docker
```

Check. 
```sh
$ docker version
Client: Docker Engine - Community
 Version:           29.1.1
 API version:       1.52
 Go version:        go1.25.4
...
 ```


Then Colima
```sh
$ brew install colima
...
```

Check if it is installed correctly.
```sh
$ colima version
colima version 0.9.1
...
```

Start it with ```docker``` container runtime. 

```sh
$ colima start --runtime docker
INFO[0000] starting colima
INFO[0000] runtime: docker
INFO[0003] creating and starting ...                     context=vm
INFO[0004] downloading disk image ...                    context=vm
INFO[0038] provisioning ...                              context=docker
INFO[0040] starting ...                                  context=docker
INFO[0042] done
```

Then check docker.

```sh
$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES

```

Let us test if we can run nginx container on colima. 

```sh
docker run --name test-nginx -p 8080:80 -d nginx
```

Open http://localhost:8080

![nginx-1](./images/nginx-1.png)

All right. 

Stop colima.
```sh
$ colima stop
```

Check.

```sh
$ colima status
FATA[0000] colima is not running
```

If we would like to change the runtime, we will first
```sh 
$ colima delete --data
```

and start colima again. 

---

### Kubernets

We will install **```kind```**.  kind is [not officially supported by Homebrew on macOS 13 Ventura](https://formulae.brew.sh/formula/kind), but it worked. 

```sh
$ brew install kind
```

Check 

```sh
$ kind version
kind v0.30.0 go1.25.4 darwin/amd64
```

We will need [```kubectl```](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/) as well.

```sh
$  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   138  100   138    0     0    916      0 --:--:-- --:--:-- --:--:--   920
100 58.9M  100 58.9M    0     0  43.9M      0  0:00:01  0:00:01 --:--:-- 53.0M

$ chmod +x kubectl
$ mv kubectl /usr/local/bin/
$ rehash
$ kubectl version
Client Version: v1.34.2
Kustomize Version: v5.7.1
Error from server (NotFound): the server could not find the requested resource
```

Let us add [autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/#enable-shell-autocompletion)
and alias in ~/.zshrc.

```sh
alias k="kubectl"
autoload -Uz compinit
compinit
source <(kubectl completion zsh)
```

Then 
```sh
$ source ~/.zshrc
```

Check if we have ```~/.kube/config```
```sh
$ ls ~/.kube/
total 16
-rw-------  1 meg  staff  5576 Dec  2 12:44 config
drwxr-x---  4 meg  staff   128 Dec  2 12:45 cache

```
Yes. 


All right. Then we will create a Kubernetes cluster. 

```sh
$ kind create cluster --name dev
Creating cluster "dev" ...
 ✓ Ensuring node image (kindest/node:v1.34.0) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-dev"
You can now use your cluster with:

kubectl cluster-info --context kind-dev

Thanks for using kind! 😊
```

Then test the cluster.

```sh
$ kubectl cluster-info --context kind-dev
Kubernetes control plane is running at https://127.0.0.1:62692
CoreDNS is running at https://127.0.0.1:62692/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

All right. 
```sh
$ k get no
NAME                STATUS   ROLES           AGE     VERSION
dev-control-plane   Ready    control-plane   8m41s   v1.34.0
```


Can we see the node on docker?
```sh
$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED       STATUS       PORTS                       NAMES
f539cb220062   kindest/node:v1.34.0   "/usr/local/bin/entr…"   6 hours ago   Up 6 hours   127.0.0.1:62692->6443/tcp   dev-control-plane
```

Wunderbar.



----
## Directory Structure

Go dose not have framework tools. Just create directories one by one. 


```sh


$ mkdir -p ./cmd/device-plugin
$ mkdir -p ./pkg/plugin
$ mkdir -p ./deployments

$ touch ./cmd/device-plugin/main.go   
...

```

Currently it looks like this.
```sh
$ tree .
.
├── GO_DEVICE_PLUGIN
├── LICENSE
├── README.md
├── cmd
│   └── device-plugin
│       └── main.go
├── deployments
│   ├── device-plugin-daemonset.yaml
│   └── test-pod.yaml
├── go.mod
├── go.sum
└── pkg
    └── plugin
        └── plugin.go
```

|              |                         |
|--------------|-------------------------|
|```go.mod``` |                           |
|```go.sum```  |                          |


---
## GoLang

### Which version?

The latest version of GoLang at the time of writing is 1.25, 
which is too new for this project. First we have to uninstall the latest
version,  

```sh
$ brew uninstall go
```

Then, install GoLang again with the version explicitly specified.  

```sh
$ brew install go@1.24
```

The non-latest version of GoLang is installed in ```/usr/local/opt```, 
while the latest version in ```/usr/local/bin```. We have to let shell know that.

Add the following in .zshrc.
```sh
export GOPATH=/usr/local/opt/go@1.24/
export PATH=$PATH:$GOPATH/bin
# export PATH="/usr/local/opt/go@1.22/bin:$PATH"
```

Then, set up ```go.mod``` at the project root.
We only need module name and go version in the initial setup.
The rest will be filled when we execute ```go mod tidy```.

```sh
module example.com/device-plugin
go 1.24.0
```

```module``` entry is the address where one can get that module. 

It **looks like** a **domainname**, but 
it is a dummy (= placeholder). One can set it up later when you 
open your project to outside, for instance 

```sh
module github.com/[YOUR GITHUB NAME]/[YOUR REPO]  
```
so that one can retrieve the module by ```go get ```.

```example.com``` can be used by anybody. 


```go 1.24``` is the version of GoLang used in this project. 



After we executed ```go mod tidy```, go.mod looks like 


```sh
module example.com/device-plugin

go 1.24.0

require (
	google.golang.org/grpc v1.77.0
	k8s.io/kubelet v0.34.2
)

require (
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
```

and ```go.sum``` is created to make sure correct dependencies (without scumming) 
used.  We do not create ```go.sum``` manually ourselves.



The latest versions of ```grpc``` and ```kubelet``` can be seen here.

- https://pkg.go.dev/k8s.io/kubelet 
- https://pkg.go.dev/google.golang.org/grpc

----

## Source Code Analysis


We would need only two GoLand codes to create this device plugin.

- ```./cmd/device-plugin/main.go```
- ```./pkg/plugin/plugin.go```

### ```main.go```

We will look at main.go first.


```go main.go
import (
	"log"
	"example.com/device-plugin/pkg/plugin"
)

func main() {
	dp := plugin.NewDevicePlugin()

	log.Println("Starting Device Plugin...")

	if err := dp.Start(); err != nil {
		log.Fatalf("Error starting plugin: %v", err)
	}

	select {}
}
```

We specified the location of this device plugin module
in ```go.mod``` like this

```sh
module example.com/device-plugin
```

This points to the project root. We have following directory 
structure 

```sh
$ tree .
.
├── GO_DEVICE_PLUGIN
├── LICENSE
├── README.md
├── cmd
│   └── device-plugin
│       └── main.go
├── deployments
│   ├── device-plugin-daemonset.yaml
│   └── test-pod.yaml
├── go.mod
├── go.sum
├── images
│   └── nginx-1.png
└── pkg
    └── plugin
        └── plugin.go

```

and ```plugin.go``` (= the main body of the plugin)
is located at ```pkg/plugin```. Therefore the ```import``` 
in main.go should be

```sh
	"example.com/device-plugin/pkg/plugin"
```

so, **[module name]** + **[folder path]**.

Note, 

```sh

	dp := plugin.NewDevicePlugin()

```

does **not** mean ```plugin``` here is the name of the diretory. 
We have three 'plugin's in  


```sh
└── pkg
    └── plugin
        └── plugin.go
```

and the first line of plugin.go

```sh
$ head -n 1 ./pkg/plugin/plugin.go
package plugin
```

The 'plugin' in ```dp := plugin.NewDevicePlugin()``` corresponds
to the third on. If ```plugin.go``` starts with 

```sh
package plugin3
```

main.go should call it like 

```sh
dp := plugin3.NewDevicePlugin()
```

In the syntax of GoLang, 

```sh

  if [definition of a parameter]; [condition] {
    [something to be executed]
  }

```

Therefor the code inside ```{}``` below will be
executed, only when the return value of dp.Start()
is not ```nil```.

```sh
	if err := dp.Start(); err != nil {
```

```select``` is 

a syntax that 　waits for
multiple channel operations without consuming CPUs.

```go
select {
  case v := <- ch1:   // input via ch1
    fmt.Println("ch1", v)
  case ch2 <- 10:    // output via ch2
    fmt.Println("Sent 10 to ch2")
}
```

If there is nothing to execute, 
```sh
select {}
```
the code waits forever (without consuming CPUs).   

This is to prevent main.go from exiting before dp.Start() returns 
a value. 

---

### ```plugin.go```

```plugin.go``` consists of 

- const
- type
- func x 7


```func``` defines methods that this device plugin can execute. 

Here we will pick one of them, ```ListAndWait```  and see it 
closely. 

```go
func (p *DevicePlugin) ListAndWatch(_ *dp.Empty,
	stream dp.DevicePlugin_ListAndWatchServer) error {
	devices := []*dp.Device{
		{ID: "dev1", Health: dp.Healthy},
	}
	return stream.Send(&dp.ListAndWatchResponse{Devices: devices})
}
```

GoLang function has the following shape. 

|                   |                            | 
|-------------------|----------------------------|
| func              | declare this is a funciton. | 
| (p *DevicePlugin) | ```p``` is a receiveer | 
| ListAndWatch      | name of function (here it is a method of DevicePlugin) | 
| (_ *dp.Empty, stream dp.DevicePlugin_ListAndWatchServer) | arguments | 
| error             | type of (optional) return value    |   
| {}　　　　　　　　　 | content of the method          | 　


A receiver (```p``` here) is like a ```self``` in 
python class definition that points to itself (= an instance 
of DevicePlugin). 

```dp``` is defined here
```go
import (
  ...
  dp "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)
```

and is an API template for kubenetes device plugin. 







---
## Test



env GOOS=linux GOARCH=amd64 go build -o device-plugin ./cmd/device-plugin


GOOS=linux GOARCH=amd64 go build -o device-plugin ./cmd/device-plugin


docker cp ./device-plugin dev-control-plane:/device-plugin
Successfully copied 15.5MB to dev-control-plane:/device-plugin

----
# END
----