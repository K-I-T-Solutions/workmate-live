package specs

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

func Probe() Specs {
	return Specs{
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Kernel: kernelVersion(),
		CPU:    cpuInfo(),
		Memory: memoryInfo(),
	}
}
func kernelVersion() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}
func cpuInfo() CPU {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return CPU{}
	}

	var model string
	cores := 0

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "model name") && model == "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				model = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "processor") {
			cores++
		}
	}

	return CPU{
		Model:   model,
		Cores:   cores,
		Threads: cores,
	}
}
func memoryInfo() Memory {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return Memory{}
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, _ := strconv.Atoi(parts[1])
				return Memory{
					TotalMB: kb / 1024,
				}
			}
		}
	}

	return Memory{}
}
