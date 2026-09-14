package gpu

import "testing"

func TestVendorFromName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// Windows, DriverDesc aus der Registry
		{"NVIDIA GeForce RTX 4070", "nvidia"},
		{"NVIDIA Quadro P2000", "nvidia"},
		{"AMD Radeon RX 7900 XTX", "amd"},
		{"Intel(R) UHD Graphics 770", "intel"},
		{"Microsoft Basic Display Adapter", "microsoft-basic"},
		// macOS, system_profiler
		{"Apple M3 Pro", "apple"},
		{"AMD Radeon Pro 5500M", "amd"},
		{"Intel Iris Plus Graphics", "intel"},
		// Unbekanntes bleibt unbekannt, statt falsch geraten zu werden
		{"Virtio GPU", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		if got := VendorFromName(tt.name); got != tt.want {
			t.Errorf("VendorFromName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
