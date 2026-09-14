package obs

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Probe sucht den OBS-Prozess in der Prozessliste des Systems.
// CreateToolhelp32Snapshot kostet wenige Millisekunden und ist damit auch
// für das Polling im Sekundentakt geeignet.
func Probe() Status {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return Status{Running: false}
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return Status{Running: false}
	}

	for {
		if IsOBSProcess(windows.UTF16ToString(entry.ExeFile[:])) {
			return Status{Running: true}
		}

		if err := windows.Process32Next(snapshot, &entry); err != nil {
			// ERROR_NO_MORE_FILES beendet die Aufzählung regulär.
			return Status{Running: false}
		}
	}
}
