//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
	"time"

	"gonesemu/nes"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth    = 256
	screenHeight   = 240
	ticksPerSecond = 60
	windowTitle    = "GoNES - NES Emulator"
)

type browserGame struct {
	system        *nes.System
	frame         *ebiten.Image
	audio         *audioOutput
	nextFPSUpdate time.Time
}

var game *browserGame
var loadROMFunction js.Func
var unloadROMFunction js.Func
var setAudioEnabledFunction js.Func
var setControllerStateFunction js.Func
var touchButtons [8]bool

func main() {
	game = &browserGame{
		frame: ebiten.NewImage(screenWidth, screenHeight),
	}

	audioOutput, err := newAudioOutput()
	if err == nil {
		game.audio = audioOutput
		defer audioOutput.Close()
	}

	loadROMFunction = js.FuncOf(loadROM)
	js.Global().Set("loadNESROM", loadROMFunction)
	unloadROMFunction = js.FuncOf(unloadROM)
	js.Global().Set("unloadNESROM", unloadROMFunction)
	setAudioEnabledFunction = js.FuncOf(setAudioEnabled)
	js.Global().Set("setNESAudioEnabled", setAudioEnabledFunction)
	setControllerStateFunction = js.FuncOf(setControllerState)
	js.Global().Set("setNESControllerState", setControllerStateFunction)

	ebiten.SetWindowTitle(windowTitle)
	ebiten.SetTPS(ticksPerSecond)
	if err := ebiten.RunGame(game); err != nil {
		reportRuntimeError(err)
	}
}

func (g *browserGame) Update() error {
	if g.system == nil {
		return nil
	}

	g.system.SetButtons1(readButtons())
	g.system.StepSeconds(1.0 / ticksPerSecond)
	return nil
}

func (g *browserGame) Draw(screen *ebiten.Image) {
	if g.system == nil {
		return
	}

	g.frame.WritePixels(g.system.Buffer().Pix)
	screen.DrawImage(g.frame, nil)

	now := time.Now()
	if !now.Before(g.nextFPSUpdate) {
		g.nextFPSUpdate = now.Add(time.Second)
		reportPerformance(ebiten.ActualFPS(), ebiten.ActualTPS())
	}
}

func (g *browserGame) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geometry ebiten.GeoM) {
	options := &ebiten.DrawImageOptions{GeoM: geometry}
	options.Filter = ebiten.FilterNearest
	screen.DrawImage(offscreen, options)
}

func (g *browserGame) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func (g *browserGame) load(data []byte) (*nes.System, error) {
	system, err := nes.NewSystemFromBytes(data)
	if err != nil {
		return nil, err
	}
	system.Reset()

	if g.audio != nil {
		system.SetAudioChannel(g.audio.Channel())
		system.SetAudioSampleRate(g.audio.SampleRate())
	}
	g.system = system
	return system, nil
}

func (g *browserGame) unload() {
	g.system = nil
	touchButtons = [8]bool{}
	g.frame.Clear()
	if g.audio != nil {
		g.audio.Flush()
	}
}

func loadROM(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeObject {
		return loadResult(fmt.Errorf("expected a Uint8Array"))
	}

	source := args[0]
	data := make([]byte, source.Get("byteLength").Int())
	if copied := js.CopyBytesToGo(data, source); copied != len(data) {
		return loadResult(fmt.Errorf("copied %d of %d ROM bytes", copied, len(data)))
	}

	system, err := game.load(data)
	if err != nil {
		return loadResult(err)
	}

	return map[string]any{
		"ok":             true,
		"mapper":         int(system.Cartridge.Mapper),
		"prgBytes":       len(system.Cartridge.PRG),
		"chrBytes":       len(system.Cartridge.CHR),
		"audioAvailable": game.audio != nil,
		"audioEnabled":   game.audio != nil && game.audio.Enabled(),
	}
}

func unloadROM(_ js.Value, _ []js.Value) any {
	game.unload()
	return nil
}

func setAudioEnabled(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeBoolean || game.audio == nil {
		return false
	}

	game.audio.SetEnabled(args[0].Bool())
	return true
}

func setControllerState(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeNumber {
		return false
	}

	mask := args[0].Int()
	for button := range touchButtons {
		touchButtons[button] = mask&(1<<button) != 0
	}
	return true
}

func loadResult(err error) map[string]any {
	return map[string]any{
		"ok":    false,
		"error": err.Error(),
	}
}

func readButtons() [8]bool {
	buttons := [8]bool{
		ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeySpace),
		ebiten.IsKeyPressed(ebiten.KeyX),
		ebiten.IsKeyPressed(ebiten.KeyShiftRight),
		ebiten.IsKeyPressed(ebiten.KeyEnter),
		ebiten.IsKeyPressed(ebiten.KeyArrowUp),
		ebiten.IsKeyPressed(ebiten.KeyArrowDown),
		ebiten.IsKeyPressed(ebiten.KeyArrowLeft),
		ebiten.IsKeyPressed(ebiten.KeyArrowRight),
	}
	for button, pressed := range touchButtons {
		buttons[button] = buttons[button] || pressed
	}
	return buttons
}

func reportRuntimeError(err error) {
	reporter := js.Global().Get("setNESRuntimeError")
	if reporter.Type() == js.TypeFunction {
		reporter.Invoke(err.Error())
	}
}

func reportPerformance(fps, tps float64) {
	reporter := js.Global().Get("setNESPerformance")
	if reporter.Type() == js.TypeFunction {
		reporter.Invoke(fps, tps)
	}
}
