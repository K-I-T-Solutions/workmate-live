package video

import (
	"sort"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Geräteklassen, unter denen Windows Kameras einsortiert: die klassische
// Image-Klasse und die seit Windows 10 genutzte Camera-Klasse.
var cameraClasses = []string{
	`SYSTEM\CurrentControlSet\Control\Class\{6bdd1fc6-810f-11d0-bec7-08002be2092f}`,
	`SYSTEM\CurrentControlSet\Control\Class\{ca3e7ab9-b4c3-4ae6-8251-579ef933890f}`,
}

// ScanDevices listet die installierten Kameras anhand ihres Anzeigenamens.
// Unter Linux liefert diese Funktion Gerätepfade (/dev/video*), hier sind es
// Namen — beides sind für das Portal nur Bezeichner.
func ScanDevices() ([]string, error) {
	seen := map[string]bool{}
	devices := []string{}

	for _, class := range cameraClasses {
		for _, name := range devicesInClass(class) {
			if seen[name] {
				continue
			}
			seen[name] = true
			devices = append(devices, name)
		}
	}

	sort.Strings(devices)
	return devices, nil
}

func devicesInClass(class string) []string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, class, registry.READ)
	if err != nil {
		return nil
	}
	defer key.Close()

	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil
	}

	var found []string

	for _, name := range subkeys {
		// Nur die durchnummerierten Instanzschlüssel (0000, 0001, …).
		if len(name) != 4 || strings.ContainsFunc(name, func(r rune) bool { return r < '0' || r > '9' }) {
			continue
		}

		sub, err := registry.OpenKey(key, name, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		desc, _, err := sub.GetStringValue("FriendlyName")
		if err != nil || desc == "" {
			desc, _, err = sub.GetStringValue("DriverDesc")
		}
		sub.Close()

		if err != nil || desc == "" {
			continue
		}

		found = append(found, strings.TrimSpace(desc))
	}

	return found
}
