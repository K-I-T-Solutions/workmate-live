package health

import (
	"os"
	"time"

	"kit.workmate/gaming-agent/internal/config"
	"kit.workmate/gaming-agent/internal/system/audio"
	"kit.workmate/gaming-agent/internal/system/gpu"
	"kit.workmate/gaming-agent/internal/system/obs"
	"kit.workmate/gaming-agent/internal/system/video"
)

func Collect(checks config.ChecksConfig) (*Status, error) {
	hostname, _ := os.Hostname()

	// Conditional probing based on config
	obsStatus := OBSStatus{}
	if checks.OBS {
		probed := obs.Probe()
		obsStatus = OBSStatus{
			Running: probed.Running,
		}
	}

	gpuStatus := gpu.Status{}
	if checks.GPU {
		gpuStatus = gpu.Probe()
	}

	audioStatus := AudioStatus{}
	if checks.Audio {
		probed := audio.Probe()
		audioStatus = AudioStatus{
			Backend: probed.Backend,
			Ready:   probed.Ready,
		}
	}

	headless := os.Getenv("DISPLAY") == "" &&
		os.Getenv("WAYLAND_DISPLAY") == ""

	videoDevices := []string{}
	if checks.Video {
		var err error
		videoDevices, err = video.ScanDevices()
		if err != nil {
			videoDevices = []string{}
		}
	}

	status := &Status{
		Timestamp: time.Now(),
		Hostname:  hostname,
		Headless:  headless,
		Video: VideoStatus{
			DeviceCount: len(videoDevices),
			Devices:     videoDevices,
		},
		OBS:   obsStatus,
		GPU:   gpuStatus,
		Audio: audioStatus,
	}

	return status, nil
}
