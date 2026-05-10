//go:build !windows

package audio

import (
	"fmt"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate = 44100
	channels   = 2
)

// Capturer handles system audio capture via PortAudio.
// On Windows it prefers a WASAPI input device (which includes loopback endpoints)
// and falls back to the default input device on other platforms.
type Capturer struct {
	device *portaudio.DeviceInfo
}

// NewCapturer initialises PortAudio and picks the best available input device.
func NewCapturer() (*Capturer, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio init: %w", err)
	}

	device, err := findBestInputDevice()
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("no usable audio input device: %w", err)
	}

	return &Capturer{device: device}, nil
}

// findBestInputDevice prefers WASAPI devices (Windows loopback-capable),
// then falls back to the system default input.
func findBestInputDevice() (*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err == nil {
		for _, d := range devices {
			if d.MaxInputChannels >= channels && d.HostApi.Type == portaudio.WASAPI {
				return d, nil
			}
		}
	}
	return portaudio.DefaultInputDevice()
}

// CaptureSnippet records audio for the given duration and returns raw PCM bytes
// (signed 16-bit little-endian, stereo, 44100 Hz).
func (c *Capturer) CaptureSnippet(duration time.Duration) ([]byte, error) {
	numSamples := int(duration.Seconds() * sampleRate)
	buffer := make([]int16, numSamples*channels)

	params := portaudio.HighLatencyParameters(c.device, nil)
	params.Input.Channels = channels
	params.SampleRate = sampleRate
	params.FramesPerBuffer = len(buffer)

	stream, err := portaudio.OpenStream(params, buffer)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return nil, fmt.Errorf("start stream: %w", err)
	}
	if err := stream.Read(); err != nil {
		return nil, fmt.Errorf("read audio: %w", err)
	}
	if err := stream.Stop(); err != nil {
		return nil, fmt.Errorf("stop stream: %w", err)
	}

	out := make([]byte, len(buffer)*2)
	for i, s := range buffer {
		out[i*2] = byte(s)
		out[i*2+1] = byte(s >> 8)
	}
	return out, nil
}

// Close releases PortAudio resources.
func (c *Capturer) Close() error {
	portaudio.Terminate()
	return nil
}
