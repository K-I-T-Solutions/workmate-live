package audio

import (
	"golang.org/x/sys/windows/registry"
)

// renderEndpoints listet die Wiedergabegeräte, die Windows kennt.
const renderEndpoints = `SOFTWARE\Microsoft\Windows\CurrentVersion\MMDevices\Audio\Render`

// deviceStateActive ist der DeviceState eines betriebsbereiten Endpunkts.
// Andere Werte stehen für deaktiviert, nicht vorhanden oder abgezogen.
const deviceStateActive = 1

// Probe meldet Audio als bereit, sobald mindestens ein aktives
// Wiedergabegerät registriert ist.
func Probe() Status {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, renderEndpoints, registry.READ)
	if err != nil {
		return Status{Backend: "wasapi", Ready: false}
	}
	defer key.Close()

	endpoints, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return Status{Backend: "wasapi", Ready: false}
	}

	for _, name := range endpoints {
		sub, err := registry.OpenKey(key, name, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		state, _, err := sub.GetIntegerValue("DeviceState")
		sub.Close()

		if err == nil && state == deviceStateActive {
			return Status{Backend: "wasapi", Ready: true}
		}
	}

	return Status{Backend: "wasapi", Ready: false}
}
