package nes

import (
	"bytes"
	"fmt"
	"image"
	"io"
)

// Hardware constants
const (
	// RAMSize is the size of the NES internal RAM in bytes (2KB)
	RAMSize = 2048

	// CPUToPPURatio defines the PPU clock rate relative to CPU (PPU runs 3x faster)
	CPUToPPURatio = 3
)

// System represents the NES emulator system containing all hardware components.
// It manages the CPU, PPU (graphics), APU (audio), memory, controllers, and ROM cartridge.
type System struct {
	CPU         *CPU
	APU         *APU
	PPU         *PPU
	Cartridge   *Cartridge
	Controller1 *Controller
	Controller2 *Controller
	Mapper      Mapper
	RAM         []byte
}

// NewSystem creates and initializes a new NES system by loading a ROM.
// The initialization order is critical: ROM → RAM → Controllers → Mapper → CPU/PPU/APU
func NewSystem(reader io.Reader) (*System, error) {
	rom, err := LoadNES(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to load ROM: %w", err)
	}

	// Initialize system components
	sys := &System{
		RAM:         initializeRAM(),
		Cartridge:   rom,
		Controller1: NewController(),
		Controller2: NewController(),
	}

	// Create mapper based on ROM type
	memMapper, err := NewMapper(sys)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapper: %w", err)
	}
	sys.Mapper = memMapper

	// Initialize CPU, APU, and PPU (order matters - they reference the system)
	sys.CPU = NewCPU(sys)
	sys.APU = NewAPU(sys)
	sys.PPU = NewPPU(sys)

	return sys, nil
}

func NewSystemFromBytes(data []byte) (*System, error) {
	return NewSystem(bytes.NewReader(data))
}

func initializeRAM() []byte {
	return make([]byte, RAMSize)
}

func (sys *System) Reset() {
	sys.CPU.Reset()
}

// Step executes one CPU instruction and synchronizes PPU and APU.
// PPU runs 3x faster than CPU, APU runs at same speed as CPU.
func (sys *System) Step() int {
	cpuCycles := sys.CPU.Step()

	sys.stepPPUCycles(cpuCycles)
	sys.stepAPUCycles(cpuCycles)

	return cpuCycles
}

func (sys *System) stepPPUCycles(cpuCycles int) {
	ppuCycles := cpuCycles * CPUToPPURatio
	for i := 0; i < ppuCycles; i++ {
		sys.PPU.Step()
	}
}

func (sys *System) stepAPUCycles(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		sys.APU.Step()
	}
}

// StepSeconds executes instructions for the specified duration.
// This maintains accurate emulation timing.
func (sys *System) StepSeconds(seconds float64) {
	totalCycles := sys.calculateCyclesForDuration(seconds)
	sys.executeCycles(totalCycles)
}

func (sys *System) calculateCyclesForDuration(seconds float64) int {
	return int(seconds * CPUFrequency)
}

func (sys *System) executeCycles(cycles int) {
	for cycles > 0 {
		cycles -= sys.Step()
	}
}

func (sys *System) Buffer() *image.RGBA {
	return sys.PPU.front
}

func (sys *System) SetButtons1(buttons [8]bool) {
	sys.Controller1.SetButtons(buttons)
}

func (sys *System) SetAudioChannel(ch chan float32) {
	sys.APU.channel = ch
}

// SetAudioSampleRate configures the audio sample rate and initializes filters.
// A sample rate of 0 disables audio.
func (sys *System) SetAudioSampleRate(sampleRate float64) {
	if sampleRate != 0 {
		sys.APU.sampleRate = CPUFrequency / sampleRate
		sys.initializeAudioFilters(sampleRate)
	} else {
		sys.APU.filterChain = nil
	}
}

func (sys *System) initializeAudioFilters(sampleRate float64) {
	// High-pass filters remove DC offset and rumble, low-pass removes high-frequency noise
	sys.APU.filterChain = FilterChain{
		HighPassFilter(float32(sampleRate), 90),
		HighPassFilter(float32(sampleRate), 440),
		LowPassFilter(float32(sampleRate), 14000),
	}
}
