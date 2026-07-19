//go:build js && wasm

package main

import (
	"encoding/binary"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	audioSampleRate = 44100
	audioBufferSize = audioSampleRate / 10
)

type audioOutput struct {
	context *audio.Context
	player  *audio.Player
	channel chan float32
	enabled bool
}

type audioStream struct {
	samples chan float32
	frame   [4]byte
	offset  int
}

func newAudioOutput() (*audioOutput, error) {
	channel := make(chan float32, audioBufferSize)
	context := audio.NewContext(audioSampleRate)
	stream := &audioStream{samples: channel, offset: len([4]byte{})}
	player, err := context.NewPlayer(stream)
	if err != nil {
		return nil, err
	}
	player.SetBufferSize(100 * time.Millisecond)
	player.Play()

	return &audioOutput{
		context: context,
		player:  player,
		channel: channel,
		enabled: true,
	}, nil
}

func (a *audioOutput) Channel() chan float32 {
	return a.channel
}

func (a *audioOutput) SampleRate() float64 {
	return float64(a.context.SampleRate())
}

func (a *audioOutput) SetEnabled(enabled bool) {
	a.enabled = enabled
	if enabled {
		a.player.SetVolume(1)
		return
	}
	a.player.SetVolume(0)
}

func (a *audioOutput) Enabled() bool {
	return a.enabled
}

func (a *audioOutput) Flush() {
	for {
		select {
		case <-a.channel:
		default:
			return
		}
	}
}

func (a *audioOutput) Close() error {
	return a.player.Close()
}

func (s *audioStream) Read(buffer []byte) (int, error) {
	for i := range buffer {
		if s.offset == len(s.frame) {
			s.loadFrame()
		}
		buffer[i] = s.frame[s.offset]
		s.offset++
	}
	return len(buffer), nil
}

func (s *audioStream) loadFrame() {
	value := float32(0)
	select {
	case value = <-s.samples:
	default:
	}
	value = max(-1, min(1, value))
	sample := uint16(int16(value * 32767))
	binary.LittleEndian.PutUint16(s.frame[0:2], sample)
	binary.LittleEndian.PutUint16(s.frame[2:4], sample)
	s.offset = 0
}
