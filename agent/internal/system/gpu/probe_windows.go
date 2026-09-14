package gpu

import (
	"sort"
	"strings"
	"sync"

	"golang.org/x/sys/windows/registry"
)

// displayClass ist die Geräteklasse der Grafikkarten in der Registry.
const displayClass = `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`

var (
	once  sync.Once
	value Status
)

// Probe liest die installierten Grafikadapter aus der Registry. Das ist
// deutlich schneller als eine WMI-Abfrage und kommt ohne Prozessstart aus.
// Das Ergebnis wird zwischengespeichert — Adapter ändern sich im laufenden
// Betrieb praktisch nie, der Health-Poller fragt aber im Sekundentakt.
func Probe() Status {
	once.Do(func() { value = probeRegistry() })
	return value
}

func probeRegistry() Status {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, displayClass, registry.READ)
	if err != nil {
		return Status{Present: false}
	}
	defer key.Close()

	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return Status{Present: false}
	}

	seenVendor := map[string]bool{}
	seenAdapter := map[string]bool{}

	var vendors, adapters []string

	for _, name := range subkeys {
		// Nur die durchnummerierten Instanzschlüssel (0000, 0001, …).
		if len(name) != 4 || strings.ContainsFunc(name, func(r rune) bool { return r < '0' || r > '9' }) {
			continue
		}

		sub, err := registry.OpenKey(key, name, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		desc, _, err := sub.GetStringValue("DriverDesc")
		sub.Close()
		if err != nil || desc == "" {
			continue
		}

		if !seenAdapter[desc] {
			seenAdapter[desc] = true
			adapters = append(adapters, desc)
		}

		if v := VendorFromName(desc); v != "" && !seenVendor[v] {
			seenVendor[v] = true
			vendors = append(vendors, v)
		}
	}

	if len(adapters) == 0 {
		return Status{Present: false}
	}

	sort.Strings(vendors)
	sort.Strings(adapters)

	return Status{
		Present:  true,
		Vendors:  vendors,
		Adapters: adapters,
	}
}
