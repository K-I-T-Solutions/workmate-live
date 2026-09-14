package gpu

type Status struct {
	Present bool     `json:"present"`
	Vendors []string `json:"vendors,omitempty"`

	// RenderNodes sind die DRM-Render-Nodes unter Linux (/dev/dri/renderD*).
	RenderNodes []string `json:"render_nodes,omitempty"`

	// Adapters sind die Namen der Grafikadapter, wie Windows und macOS sie
	// melden. Unter Linux bleibt das Feld leer.
	Adapters []string `json:"adapters,omitempty"`
}
