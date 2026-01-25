package gpu

type Status struct {
	Present     bool     `json:"present"`
	Vendors     []string `json:"vendors,omitempty"`
	RenderNodes []string `json:"render_nodes,omitempty"`
}
