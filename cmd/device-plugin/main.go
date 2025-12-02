package main

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
