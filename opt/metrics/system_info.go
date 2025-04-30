package main

import (
	"os"
    "os/user"
	"log"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func getSystemMetrics() (string, error) {
    // Get CPU usage percentage
    cpuPercentages, err := cpu.Percent(0, false)
    if err != nil {
        return "", err
    }

    // Get disk usage
    diskUsage, err := disk.Usage("/")
    if err != nil {
        return "", err
    }

    // Get the hostname
    hostname, err := os.Hostname()
    if err != nil {
        return "", err
    }

	hostID, _ := host.HostID()

	memoryUsage, _ := mem.VirtualMemory()

    // Format the metrics into a string
    metrics := fmt.Sprintf(
        "Metrics for hostname %s:\n\nCPU Usage: %.2f%%\nDisk Usage: %.2f%%\nMem Usage: %.2f%%\n" + 
        "Disk Info - Total: %d bytes, Free: %d bytes, Used: %d bytes\n" + 
		"Mem Info - Total: %d bytes, Free: %d bytes, Used: %d bytes\n" +
        "Host Id: %s\n",
        hostname, cpuPercentages[0], diskUsage.UsedPercent, memoryUsage.UsedPercent,
        diskUsage.Total, diskUsage.Free, diskUsage.Used, 
		memoryUsage.Total, memoryUsage.Free, memoryUsage.Used,
        hostID,
    )

    return metrics, nil
}

func getSystemInfo() map[string]string {
    result := make(map[string]string)
    user, err := user.Current()
    hostname, _ := os.Hostname()
    hostID, _ := host.HostID()
	if err != nil {
		log.Print("Error retrieving system info: ", err.Error())
	}

	username := user.Username

    result["guid"] = hostID
    result["hostname"] = hostname
    result["username"] = username
    
    return result
}
