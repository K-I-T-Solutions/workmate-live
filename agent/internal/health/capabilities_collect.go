package health

func CollectCapabilities(status *Status) Capabilities {
	canVideo := status.Video.DeviceCount > 0
	canAudio := status.Audio.Ready

	canStream := canVideo && canAudio && !status.Headless && status.OBS.Running

	return Capabilities{
		CanVideo:  canVideo,
		CanAudio:  canAudio,
		CanStream: canStream,
	}
}
