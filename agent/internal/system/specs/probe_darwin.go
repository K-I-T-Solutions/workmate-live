package specs

import (
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
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

// kernelVersion liefert die Darwin-Version, etwa "24.1.0".
// sysctl ist ein direkter Kernel-Aufruf und damit auch beim Polling im
// Sekundentakt unbedenklich — anders als ein system_profiler-Prozess.
func kernelVersion() string {
	release, err := unix.Sysctl("kern.osrelease")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(release)
}

func cpuInfo() CPU {
	model, err := unix.Sysctl("machdep.cpu.brand_string")
	if err != nil {
		model = ""
	}

	cores, err := unix.SysctlUint32("hw.physicalcpu")
	if err != nil {
		cores = 0
	}

	threads, err := unix.SysctlUint32("hw.logicalcpu")
	if err != nil {
		threads = uint32(runtime.NumCPU())
	}

	return CPU{
		Model:   strings.TrimSpace(model),
		Cores:   int(cores),
		Threads: int(threads),
	}
}

func memoryInfo() Memory {
	total, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return Memory{}
	}

	return Memory{TotalMB: int(total / (1024 * 1024))}
}
