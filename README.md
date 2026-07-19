# GoNES Browser Emulator

A Nintendo Entertainment System emulator written in Go and delivered to browsers through WebAssembly.

**[Play GoNES in your browser](https://hotpath-hooligan.github.io/nes-emu/)**

## Screenshots

<p align="center">
  <img src="screenshots/Screenshot from 2025-10-20 23-52-56.png" width="32%" alt="Game Screenshot 1">
  <img src="screenshots/Screenshot from 2025-10-20 23-53-38.png" width="32%" alt="Game Screenshot 2">
  <img src="screenshots/Screenshot from 2025-10-20 23-53-56.png" width="32%" alt="Game Screenshot 3">
</p>

## Features

- Complete emulation path for the MOS 6502 CPU, PPU graphics, five-channel APU audio, memory, cartridges, and controllers
- Runs entirely in the browser as WebAssembly; locally selected ROMs stay on your device
- Ebitengine-powered 60 Hz game loop, rendering, keyboard/touch input, and audio output
- Crisp nearest-neighbor scaling from the NES's native 256 × 240 resolution
- Optional vintage CRT display with scanlines, a subtle color mask, increased contrast, and a screen-edge vignette
- Responsive on-screen controls and an optional FPS/TPS performance overlay
- Automatically generated launcher for bundled ROMs

## Cartridge mapper support

NES cartridges use mapper hardware to bank-switch program and graphics data beyond the console's directly addressable memory. GoNES currently implements six commonly used mapper families:

| Mapper | Board family | What it provides |
| --- | --- | --- |
| 0 | NROM | The original fixed 16 KB or 32 KB program-ROM layout |
| 1 | MMC1 | Configurable program/graphics banking and mirroring |
| 2 | UxROM | A switchable 16 KB program bank alongside a fixed bank |
| 3 | CNROM | Switchable graphics-ROM banks |
| 7 | AxROM | Switchable 32 KB program banks with one-screen mirroring |
| 11 | Color Dreams | Combined program-ROM and graphics-ROM bank switching |

Mapper support describes the cartridge hardware GoNES understands. Individual games can still depend on timing behavior or board variants that are not fully emulated yet.

## Run in a browser

Prerequisites:

- Go 1.22 or newer
- Python 3, or another local static file server

Build the WebAssembly module and start the local server:

```bash
git clone https://github.com/hotpath-hooligan/nes-emu.git
cd nes-emu
mkdir -p roms
# Add your legally obtained .nes cartridge dumps to roms/
make serve
```

Open `http://localhost:8080`, select a bundled game or choose a legally obtained `.nes` file from your device, and click the game canvas. The ROM stays in the browser.

Every tracked `.nes` file directly inside `roms/` is copied into the generated bundle and listed in the launcher automatically. Filenames such as `Mega Man 2 (USA).nes` are cleaned up for display without renaming the source files. A deployed ROM is publicly downloadable, so only commit files you have the right to distribute.

## Controls

| Key | Function |
| --- | --- |
| Arrow keys | D-pad |
| Z or Space | A button |
| X | B button |
| Enter | Start |
| Right Shift | Select |
| Escape | Power off the current game and return to the ROM chooser |

## Commands

```bash
make build                         # Build the minimal release in dist/
make serve                         # Build and serve dist/ on port 8080
make romInfo ROM_FILE=game.nes     # Display ROM header information
make clean                         # Remove generated browser files
make help                          # List available commands
```

`make web` and `make web-serve` remain as aliases for `make build` and `make serve`.

## Project structure

```text
nes-emu/
├── nes/                 # Platform-neutral emulator core
├── web/
│   ├── index.html       # Browser interface and WASM bootstrap
│   ├── main_wasm.go     # Browser game loop, rendering, and input
│   └── audio_wasm.go    # Browser audio output
├── roms/               # Cartridge files bundled into the deployment
├── tools/romcatalog/    # Bundled-ROM copier and catalog generator
├── dist/                # Generated deployable files only
├── Makefile             # Browser build and local server commands
└── README.md
```

## Client bundle

`make build` recreates `dist/` from an explicit allowlist. Deploy only this directory:

| File | When downloaded | Purpose |
| --- | --- | --- |
| `index.html` | Initial page | Inline UI, styles, and bootstrap code |
| `favicon.svg` | Initial page | NES controller tab icon |
| `wasm_exec.js` | Initial page | Go's required browser runtime |
| `main.wasm` | Initial page | Emulator core, Ebitengine browser runtime, rendering, input, and audio code |
| `roms/catalog.json` | Initial page | Generated list of bundled games |
| `roms/*.nes` | Only when its game button is clicked | User-supplied cartridge dumps |

Go source, native OpenGL/PortAudio code, screenshots, documentation, and reference material are not copied to `dist/`. The build strips debug symbols and local paths, reports raw and gzip sizes, and fails if compressed WASM grows beyond 3 MB. Production hosting should enable gzip or Brotli for `.wasm` and `.js` responses.

## Technical details

- CPU: MOS 6502 at 1.79 MHz (NTSC)
- Resolution: 256 × 240 pixels
- Frame rate: 60 FPS (NTSC)
- Audio core: five NES channels at 44.1 kHz
- Rendering: Ebitengine compiled for `js/wasm`

PAL timing and some advanced mapper behaviors are not yet supported.

## Roadmap

- Add more cartridge mappers to expand game compatibility
- Improve audio timing and buffering to eliminate intermittent chopping and dropouts
- Add save-state and load-state support so a game can be resumed in the browser
- Continue improving timing accuracy and compatibility across the CPU, PPU, and APU

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

Only publish ROMs you have the right to distribute. For local play, use homebrew or public-domain ROMs, or personal cartridge backups made and used in accordance with applicable law.
