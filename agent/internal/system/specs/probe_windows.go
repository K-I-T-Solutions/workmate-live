package specs

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func Probe() Specs {
	return Specs{
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Kernel: osVersion(),
		CPU:    cpuInfo(),
		Memory: memoryInfo(),
	}
}

// Statische Angaben einmal ermitteln — sie ändern sich zur Laufzeit nicht,
// und der Health-Poller fragt sie im Sekundentakt ab.
var (
	osOnce  sync.Once
	osValue string

	cpuOnce  sync.Once
	cpuValue CPU
)

// osVersion liefert die Windows-Version als "10.0.22631 (23H2)".
func osVersion() string {
	osOnce.Do(func() {
		v := windows.RtlGetVersion()
		osValue = fmt.Sprintf("%d.%d.%d", v.MajorVersion, v.MinorVersion, v.BuildNumber)

		// Die Anzeigeversion (22H2, 23H2, …) steht nur in der Registry.
		key, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
		if err != nil {
			return
		}
		defer key.Close()

		if display, _, err := key.GetStringValue("DisplayVersion"); err == nil && display != "" {
			osValue += " (" + display + ")"
		}
	})

	return osValue
}

// cpuInfo liest Modell und Kernzahl aus der Registry. Jeder Prozessor hat dort
// einen eigenen Unterschlüssel, deren Anzahl entspricht den logischen Kernen.
func cpuInfo() CPU {
	cpuOnce.Do(func() {
		threads := runtime.NumCPU()
		cpuValue = CPU{Cores: threads, Threads: threads}

		key, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
		if err != nil {
			return
		}
		defer key.Close()

		if name, _, err := key.GetStringValue("ProcessorNameString"); err == nil {
			cpuValue.Model = strings.TrimSpace(name)
		}
	})

	return cpuValue
}

// memoryStatusEx entspricht MEMORYSTATUSEX aus der Windows-API.
// x/sys/windows bringt dafür keine Definition mit.
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

var procGlobalMemoryStatusEx = windows.NewLazySystemDLL("kernel32.dll").
	NewProc("GlobalMemoryStatusEx")

// memoryInfo fragt den installierten Arbeitsspeicher ab.
func memoryInfo() Memory {
	var status memoryStatusEx
	status.Length = uint32(unsafe.Sizeof(status))

	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 {
		return Memory{}
	}

	return Memory{TotalMB: int(status.TotalPhys / (1024 * 1024))}
}
