package gpu

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Probe() Status {
	renderNodes, _ := filepath.Glob("/dev/dri/renderD*")
	if len(renderNodes) == 0 {
		return Status{Present: false}
	}

	vendors := detectVendors()

	sort.Strings(renderNodes)
	sort.Strings(vendors)

	return Status{
		Present:     true,
		Vendors:     vendors,
		RenderNodes: renderNodes,
	}
}

func detectVendors() []string {
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return nil
	}

	seen := map[string]bool{}
	var vendors []string

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "card") {
			continue
		}

		vendorFile := filepath.Join("/sys/class/drm", e.Name(), "device/vendor")
		data, err := os.ReadFile(vendorFile)
		if err != nil {
			continue
		}

		vendor := strings.TrimSpace(string(data))
		name := vendorName(vendor)
		if name == "" || seen[name] {
			continue
		}

		seen[name] = true
		vendors = append(vendors, name)
	}

	return vendors
}

func vendorName(id string) string {
	switch id {
	case "0x8086":
		return "intel"
	case "0x1002":
		return "amd"
	case "0x10de":
		return "nvidia"
	default:
		return "unknown"
	}
}
