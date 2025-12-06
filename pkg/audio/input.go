package audio

import (
	"unsafe"

	"github.com/gen2brain/malgo"
)

type AudioInput struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
	Frames chan []float32
}

func NewAudioInput(sampleRate uint32) (*AudioInput, error) {
	ai := &AudioInput{
		Frames: make(chan []float32, 8),
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	ai.ctx = ctx

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = sampleRate

	callbacks := malgo.DeviceCallbacks{
		Data: func(_, inputSamples []byte, frameCount uint32) {
			// convert bytes to float32 using unsafe slice
			if len(inputSamples) > 0 {
				samplesF32 := unsafe.Slice((*float32)(unsafe.Pointer(&inputSamples[0])), frameCount)
				// make a copy to avoid data race
				buf := make([]float32, frameCount)
				copy(buf, samplesF32)
				ai.Frames <- buf
			}
		},
	}

	dev, err := malgo.InitDevice(ctx.Context, deviceConfig, callbacks)
	if err != nil {
		return nil, err
	}
	ai.device = dev

	return ai, nil
}

func (ai *AudioInput) Start() error {
	return ai.device.Start()
}

func (ai *AudioInput) Stop() {
	ai.device.Stop()
	ai.device.Uninit()
	ai.ctx.Uninit()
	close(ai.Frames)
}
