# GoNES Browser Emulator

A Nintendo Entertainment System emulator delivered to browsers through WebAssembly.

**[Play GoNES in your browser](https://hotpath-hooligan.github.io/nes-emu/)**

## Screenshots

<p align="center">
  <img src="screenshots/Screenshot from 2025-10-20 23-52-56.png" width="32%" alt="Game Screenshot 1">
  <img src="screenshots/Screenshot from 2025-10-20 23-53-38.png" width="32%" alt="Game Screenshot 2">
  <img src="screenshots/Screenshot from 2025-10-20 23-53-56.png" width="32%" alt="Game Screenshot 3">
</p>

## Features

- Complete emulation path for the MOS 6502 CPU, PPU graphics, five-channel APU audio, memory, cartridges, and controllers
- Runs entirely in the browser as WebAssembly; completely offline
- Ebitengine-powered 60 Hz game loop, rendering, keyboard/touch input, and audio output
- Crisp nearest-neighbor scaling from the NES's native 256 × 240 resolution
- Optional Vintage CRT display to feel the nostalgia!
- Bundled ROMs

## Cartridge mapper support

GoNES currently implements six commonly used mapper families:

| Mapper | Board family | What it provides |
| --- | --- | --- |
| 0 | NROM | The original fixed 16 KB or 32 KB program-ROM layout |
| 1 | MMC1 | Configurable program/graphics banking and mirroring |
| 2 | UxROM | A switchable 16 KB program bank alongside a fixed bank |
| 3 | CNROM | Switchable graphics-ROM banks |
| 7 | AxROM | Switchable 32 KB program banks with one-screen mirroring |
| 11 | Color Dreams | Combined program-ROM and graphics-ROM bank switching |

Note: Mapper support describes the cartridge hardware GoNES understands. Individual games can still depend on timing behavior or board variants that are not fully emulated yet.

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
