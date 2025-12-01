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

We will build the architecture in the following order.

1. Lima
2. Colima
3. kind
4. GoLang


First Lima.
```sh
$ brew install lima
```
Start it.

```sh 
$ limactl start
```

Then Colima
```sh
$ brew install colima
```
Start it.

```sh 
$ colima start
```

Let us test if we can run nginx container on colima. 

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