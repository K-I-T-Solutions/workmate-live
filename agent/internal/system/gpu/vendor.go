package gpu

import "strings"

// VendorFromName leitet den Hersteller aus dem Adapternamen ab, wie ihn
// Windows (Registry) und macOS (system_profiler) melden. Linux nutzt
// stattdessen die PCI-Vendor-ID und braucht diese Funktion nicht.
func VendorFromName(name string) string {
	n := strings.ToLower(name)

	switch {
	case strings.Contains(n, "nvidia"), strings.Contains(n, "geforce"),
		strings.Contains(n, "quadro"), strings.Contains(n, "rtx"):
		return "nvidia"
	case strings.Contains(n, "amd"), strings.Contains(n, "radeon"):
		return "amd"
	case strings.Contains(n, "intel"):
		return "intel"
	case strings.Contains(n, "apple"):
		return "apple"
	case strings.Contains(n, "microsoft basic"):
		return "microsoft-basic"
	}

	return "unknown"
}
