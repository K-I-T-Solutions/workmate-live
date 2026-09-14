package gpu

import (
	"context"
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	once  sync.Once
	value Status
)

// Probe liest die Grafikadapter über system_profiler.
//
// Der Aufruf dauert rund eine Sekunde, deshalb wird das Ergebnis
// zwischengespeichert: Grafikkarten ändern sich im laufenden Betrieb nicht,
// der Health-Poller fragt aber im Sekundentakt.
func Probe() Status {
	once.Do(func() { value = probeSystemProfiler() })
	return value
}

// displaysReport bildet den Teil der system_profiler-Ausgabe ab, den wir brauchen.
type displaysReport struct {
	Displays []struct {
		Name   string `json:"_name"`
		Vendor string `json:"spdisplays_vendor"`
	} `json:"SPDisplaysDataType"`
}

func probeSystemProfiler() Status {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "system_profiler", "-json", "SPDisplaysDataType").Output()
	if err != nil {
		return Status{Present: false}
	}

	var report displaysReport
	if err := json.Unmarshal(out, &report); err != nil {
		return Status{Present: false}
	}

	seenVendor := map[string]bool{}
	var vendors, adapters []string

	for _, d := range report.Displays {
		name := strings.TrimSpace(d.Name)
		if name == "" {
			continue
		}
		adapters = append(adapters, name)

		if v := vendorName(d.Vendor, name); v != "" && !seenVendor[v] {
			seenVendor[v] = true
			vendors = append(vendors, v)
		}
	}

	if len(adapters) == 0 {
		return Status{Present: false}
	}

	sort.Strings(vendors)
	sort.Strings(adapters)

	return Status{Present: true, Vendors: vendors, Adapters: adapters}
}

// vendorName zieht Vendor-Feld und Adapternamen zusammen, weil system_profiler
// den Hersteller nicht immer eigenständig ausweist.
func vendorName(vendor, name string) string {
	return VendorFromName(vendor + " " + name)
}
