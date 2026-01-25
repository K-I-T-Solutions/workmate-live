package specs

type Specs struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Kernel string `json:"kernel"`

	CPU    CPU    `json:"cpu"`
	Memory Memory `json:"memory"`
}

type CPU struct {
	Model   string `json:"model"`
	Cores   int    `json:"cores"`
	Threads int    `json:"threads"`
}

type Memory struct {
	TotalMB int `json:"total_mb"`
}
