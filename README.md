# GoNES Browser Emulator

A Nintendo Entertainment System emulator written in Go and delivered to browsers through WebAssembly.

## Screenshots

<p align="center">
  <img src="screenshots/Screenshot from 2025-10-20 23-52-56.png" width="32%" alt="Game Screenshot 1">
  <img src="screenshots/Screenshot from 2025-10-20 23-53-38.png" width="32%" alt="Game Screenshot 2">
  <img src="screenshots/Screenshot from 2025-10-20 23-53-56.png" width="32%" alt="Game Screenshot 3">
</p>

## Features

- 6502 CPU, PPU, APU, memory, and controller emulation
- Browser rendering and keyboard input through Ebitengine
- Pixel-perfect nearest-neighbor scaling
- Automatically generated bundled-ROM library
- Local ROM selection without uploading ROM data
- Mapper 0, 1, 2, 3, 7, and 11 support

## Run in a browser

Prerequisites:

- Go 1.22 or newer
- Python 3, or another local static file server

Build the WebAssembly module and start the local server:

```bash
git clone https://github.com/akap-hub/go-nes.git
cd go-nes
mkdir -p roms
# Add your legally obtained .nes cartridge dumps to roms/
make serve
```

Open `http://localhost:8080`, select a bundled game or choose a legally obtained `.nes` file from your device, and click the game canvas. The ROM stays in the browser.

Every `.nes` file directly inside `roms/` is copied into the generated bundle and listed in the launcher automatically. Filenames such as `Mega Man 2 (USA).nes` are cleaned up for display without renaming the source files. ROM files are ignored by Git; supply your own legally obtained cartridge dumps.

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
go-nes/
├── nes/                 # Platform-neutral emulator core
├── web/
│   ├── index.html       # Browser interface and WASM bootstrap
│   ├── main_wasm.go     # Browser game loop, rendering, and input
│   └── audio_wasm.go    # Browser audio output
├── roms/               # Local, Git-ignored cartridge dumps
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
| `wasm_exec.js` | Initial page | Go's required browser runtime |
| `main.wasm` | Initial page | Emulator core, Ebitengine browser runtime, rendering, input, and audio code |
| `roms/catalog.json` | Initial page | Generated list of bundled games |
| `roms/*.nes` | Only when its game button is clicked | User-supplied cartridge dumps |

Go source, native OpenGL/PortAudio code, screenshots, documentation, and reference material are not copied to `dist/`. The build strips debug symbols and local paths, reports raw and gzip sizes, and fails if compressed WASM grows beyond 3 MB. Production hosting should enable gzip or Brotli for `.wasm` and `.js` responses.

## GitHub Pages deployment

Every push to `master` builds `dist/` and deploys it through the `github-pages` environment. The workflow can also be run manually from the Actions tab.

Before the first deployment, set **Settings → Pages → Build and deployment → Source** to **GitHub Actions**. No deployment branch or committed build output is required.

## Technical details

- CPU: MOS 6502 at 1.79 MHz (NTSC)
- Resolution: 256 × 240 pixels
- Frame rate: 60 FPS (NTSC)
- Audio core: five NES channels at 44.1 kHz
- Rendering: Ebitengine compiled for `js/wasm`

PAL timing, some advanced mapper features, and browser save states are not yet supported.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

Only use homebrew or public-domain ROMs, or personal cartridge backups made and used in accordance with applicable law. Commercial ROMs are not tracked by this repository.
