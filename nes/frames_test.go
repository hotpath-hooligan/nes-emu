package nes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestFrameHashes prints a hash of the rendered framebuffer plus CPU/PPU state
// at several points during emulation. It is a golden-value check used to prove
// that optimization work does not change emulated behavior.
func TestFrameHashes(t *testing.T) {
	roms, err := filepath.Glob("../roms/*.nes")
	if err != nil || len(roms) == 0 {
		t.Skip("no roms available")
	}
	for _, rom := range roms {
		data, err := os.ReadFile(rom)
		if err != nil {
			t.Fatal(err)
		}
		sys, err := NewSystemFromBytes(data)
		if err != nil {
			t.Fatal(err)
		}
		sys.Reset()
		samples := make(chan float32, 1<<16)
		sys.SetAudioChannel(samples)
		sys.SetAudioSampleRate(44100)

		h := sha256.New()
		for frame := 0; frame < 600; frame++ {
			sys.StepSeconds(1.0 / 60.0)
			if frame%60 == 59 {
				h.Write(sys.Buffer().Pix)
				fmt.Fprintf(h, "%d/%d/%d/%d/%d/%d",
					sys.CPU.PC, sys.CPU.A, sys.CPU.X, sys.CPU.Y, sys.CPU.SP, sys.CPU.Cycles)
				fmt.Fprintf(h, "|%d/%d/%d/%d", sys.PPU.Frame, sys.PPU.ScanLine, sys.PPU.Cycle, sys.PPU.v)
			}
			for len(samples) > 0 {
				fmt.Fprintf(h, "%f", <-samples)
			}
		}
		t.Logf("%-40s %s", filepath.Base(rom), hex.EncodeToString(h.Sum(nil))[:32])
	}
}
