package malbeep

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/faiface/beep"
	"github.com/gen2brain/malgo"
)

// Basic input audio device.
type Sink struct {
	device *malgo.Device
	buf    [][2]float64
	mixer  beep.Mixer
}

// Creates new sink. Automatically inits malgo context if needed.
func NewSink(sampleRate uint32) (*Sink, error) {
	sink := &Sink{}

	config := malgo.DefaultDeviceConfig(malgo.Playback)
	config.Playback.Format = malgo.FormatS16
	config.Playback.Channels = 2
	config.SampleRate = sampleRate
	config.Alsa.NoMMap = 1

	callbacks := malgo.DeviceCallbacks{}
	callbacks.Data = func(output, input []byte, frames uint32) {
		sink.write(output, frames)
	}

	ctx, err := initContext()
	if err != nil {
		return nil, err
	}

	sink.device, err = malgo.InitDevice(ctx.Context, config, callbacks)
	if err != nil {
		return nil, err
	}

	if err = sink.device.Start(); err != nil {
		return nil, err
	}

	return sink, nil
}

// Returns device sample rate.
func (sink *Sink) SampleRate() beep.SampleRate {
    return beep.SampleRate(sink.device.SampleRate())
}

// Plays beep streams.
func (sink *Sink) Play(s ...beep.Streamer) {
	sink.mixer.Add(s...)
}

// Closes device. Returns error when device is already closed. Frees malgo context if that was the last device using it.
func (sink *Sink) Close() error {
	if sink.device == nil {
		return fmt.Errorf("sink is already closed")
	}
	sink.device.Uninit()
	sink.device = nil
	freeContext()
	return nil
}

func (sink *Sink) write(data []byte, frames uint32) {
	if sink.buf == nil || uint32(len(sink.buf)) < frames {
		sink.buf = make([][2]float64, frames)
	}
	sink.mixer.Stream(sink.buf)

	input := unsafe.Slice((*float64)(unsafe.Pointer(&sink.buf[0][0])), frames*2)
	output := unsafe.Slice((*int16)(unsafe.Pointer(&data[0])), frames*2)

	for i, value := range input {
		value = math.Min(math.Max(value, -1), 1)
		output[i] = int16(value * (1<<15 - 1))
	}
}
