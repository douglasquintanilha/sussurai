package main

import (
	"fmt"
	"math"
	"sync"

	"github.com/gen2brain/malgo"
)

const (
	whisperSampleRate = 16000
	recordChannels    = 1
)

type Recorder struct {
	ctx       *malgo.AllocatedContext
	device    *malgo.Device
	mu        sync.Mutex
	samples   []float32
	recording bool
}

func NewRecorder() (*Recorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("malgo init: %w", err)
	}
	return &Recorder{ctx: ctx}, nil
}

func (r *Recorder) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.device != nil {
		r.device.Uninit()
		r.device = nil
	}
	_ = r.ctx.Uninit()
	r.ctx.Free()
}

func (r *Recorder) Start() error {
	r.mu.Lock()
	if r.device != nil {
		r.device.Stop()
		r.device.Uninit()
		r.device = nil
	}
	r.samples = r.samples[:0]
	r.recording = true
	r.mu.Unlock()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = recordChannels
	deviceConfig.SampleRate = whisperSampleRate

	onData := func(outputSamples, inputSamples []byte, frameCount uint32) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if !r.recording {
			return
		}
		for i := 0; i+3 < len(inputSamples); i += 4 {
			bits := uint32(inputSamples[i]) |
				uint32(inputSamples[i+1])<<8 |
				uint32(inputSamples[i+2])<<16 |
				uint32(inputSamples[i+3])<<24
			r.samples = append(r.samples, math.Float32frombits(bits))
		}
	}

	device, err := malgo.InitDevice(r.ctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onData})
	if err != nil {
		return fmt.Errorf("init capture device: %w", err)
	}

	r.mu.Lock()
	r.device = device
	r.mu.Unlock()

	if err := device.Start(); err != nil {
		return fmt.Errorf("start capture: %w", err)
	}
	return nil
}

func (r *Recorder) Stop() ([]float32, float64, error) {
	r.mu.Lock()
	r.recording = false
	raw := make([]float32, len(r.samples))
	copy(raw, r.samples)
	dev := r.device
	r.device = nil
	r.mu.Unlock()

	if dev != nil {
		dev.Stop()
		dev.Uninit()
	}

	if len(raw) == 0 {
		return nil, 0, fmt.Errorf("no audio recorded")
	}

	duration := float64(len(raw)) / float64(whisperSampleRate)
	return raw, duration, nil
}
