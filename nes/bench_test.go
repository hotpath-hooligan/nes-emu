package nes

import (
	"os"
	"testing"
)

func loadBenchSystem(b *testing.B, path string) *System {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("rom unavailable: %v", err)
	}
	sys, err := NewSystemFromBytes(data)
	if err != nil {
		b.Fatal(err)
	}
	sys.Reset()
	sys.SetAudioChannel(make(chan float32, 4096))
	sys.SetAudioSampleRate(44100)
	// Run a few seconds so the game reaches actual gameplay rendering.
	for i := 0; i < 180; i++ {
		sys.StepSeconds(1.0 / 60.0)
	}
	return sys
}

func benchmarkFrames(b *testing.B, path string) {
	sys := loadBenchSystem(b, path)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.StepSeconds(1.0 / 60.0)
	}
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds()/60.0, "x-realtime")
}

func BenchmarkSMB(b *testing.B)         { benchmarkFrames(b, "../roms/Super_mario_brothers.nes") }
func BenchmarkContra(b *testing.B)      { benchmarkFrames(b, "../roms/Contra (USA).nes") }
func BenchmarkMegaMan2(b *testing.B)    { benchmarkFrames(b, "../roms/Mega Man 2 (USA).nes") }
func BenchmarkBattletoads(b *testing.B) { benchmarkFrames(b, "../roms/Battletoads (USA).nes") }
