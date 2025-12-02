# GO Device Plugin 
Test device plugin with GoLang.

## Table of Contents



## Motivation

First experiment to understand how a device plugin works with Kubernetes. We will write one in GoLang. 


---
## Architecture

### Kind

As my laptop is recently overhauled (the system is reinstalled 
    from the operation system up), we have a complete free hand to select the tech-stuck. As a Kubernetes backbone, I would have following options. 

    - native kubernetes with kubeadm  
    - microk8s
    - minikube
    - kind
    - k3s

This time I opt to kind, as my laptop is now under performant (= hardware is old) for the tasks it should do.  


### Container Runtime

As docker desktop is too heavy for my machine, we will go for light-weight docker desktop alternative, [Colima](https://github.com/abiosoft/colima).

### VM

Therefore we will create backbone VMs with [Lima](https://github.com/lima-vm/lima).


---
## Infrastructure

### Container runtime

We will build the architecture in the following order.

1. Lima
2. Colima
3. docker (without 'desktop') 
4. kind
5. GoLang

Note, I could not use ```nerdctl``` as it does not support macOS 13 ventura.
```sh

$ brew install nerdctl
Warning: You are using macOS 13.
We (and Apple) do not provide support for this old version.
You may have better luck with MacPorts which supports older versions of macOS:
  https://www.macports.org
...
```

Roughly speaking,

|             |                   |
|-------------|-------------------|  
| docker      | docker            |
| Colima      | ~ docker desktop  |
| lima        | ~ vagrant         | 
| QEMU        | ~ virtual box     |


Okay, first lima,
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
Note, I wanted to use ```containerd```, but containerd requires qemu, and qemu 
dose not run on macOS 13 Ventura. Therefore I have to live with docker. Sigh... 

Test docker, 
```sh
docker run --name test-nginx -p 8080:80 -d nginx
```

Open http://localhost:8080

[nginx-1](./images/nginx-1.png)



Stop it. 
```sh
$ colima stop
```

Check.

```sh
$ colima status
FATA[0000] colima is not running
```

Let us test if we can run nginx container on colima. 


### Kubernets

We will install ```kind```.  kind is [not officially supported by Homebrew on macOS 13 Ventura](https://formulae.brew.sh/formula/kind), 
but it worked. 

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

Then test eth cluster.

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
## Source Code 

Write 

- ./cmd/device-plugin/main.go
- ./pkg/plugin/plugin.go




----
# END
----