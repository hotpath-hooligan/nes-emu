# NES Memory Layout Documentation

## Overview
The Nintendo Entertainment System (NES) uses a 6502-based CPU with a 16-bit address bus, allowing it to address 64KB of memory. However, the actual memory layout is more complex due to memory mapping, mirroring, and separate address spaces for the CPU and PPU.

## CPU Memory Map (64KB Address Space)

```
Address Range    Size     Description
--------------------------------------------------------------------------------
$0000-$07FF      2KB      Internal RAM
$0800-$0FFF      -        Mirror of $0000-$07FF
$1000-$17FF      -        Mirror of $0000-$07FF
$1800-$1FFF      -        Mirror of $0000-$07FF

$2000-$2007      8B       PPU Registers
$2008-$3FFF      -        Mirrors of $2000-$2007 (repeats every 8 bytes)

$4000-$4017      24B      APU and I/O Registers
$4018-$401F      -        APU and I/O functionality (disabled on most systems)

$4020-$5FFF      ~8KB     Expansion ROM (rarely used)

$6000-$7FFF      8KB      SRAM (Battery-backed save RAM on cartridge)
$8000-$FFFF      32KB     PRG-ROM (Program ROM from cartridge)
```

## Internal RAM ($0000-$07FF)

The NES has 2KB of internal RAM which is mirrored 4 times in the address space:
- $0000-$00FF: Zero Page (fast access for 6502)
- $0100-$01FF: Stack (6502 stack, grows downward)
- $0200-$02FF: Often used for sprite data (OAM buffer)
- $0300-$07FF: General purpose RAM

**Key Concept**: Writing to $0000 is the same as writing to $0800, $1000, or $1800 due to mirroring.

## PPU Register Map ($2000-$2007)

These registers control the Picture Processing Unit:

| Address | Register | Access | Description |
|---------|----------|--------|-------------|
| $2000   | PPUCTRL  | Write  | PPU control flags (NMI enable, sprite/bg patterns, etc.) |
| $2001   | PPUMASK  | Write  | PPU mask (enable rendering, color emphasis) |
| $2002   | PPUSTATUS| Read   | PPU status (VBlank, sprite 0 hit, sprite overflow) |
| $2003   | OAMADDR  | Write  | OAM address port |
| $2004   | OAMDATA  | R/W    | OAM data port (sprite attributes) |
| $2005   | PPUSCROLL| Write  | Scroll position (write twice: X, then Y) |
| $2006   | PPUADDR  | Write  | PPU address (write twice: high byte, then low byte) |
| $2007   | PPUDATA  | R/W    | PPU data port (reads/writes to VRAM) |

**Important**: Registers $2000-$2007 are mirrored throughout $2008-$3FFF. The PPU only decodes the bottom 3 bits of the address.

## APU and I/O Registers ($4000-$4017)

| Address | Register | Description |
|---------|----------|-------------|
| $4000-$4003 | Pulse 1 | Square wave channel 1 (melody) |
| $4004-$4007 | Pulse 2 | Square wave channel 2 (harmony) |
| $4008-$400B | Triangle | Triangle wave channel (bass) |
| $400C-$400F | Noise   | Noise channel (percussion/effects) |
| $4010-$4013 | DMC     | Delta Modulation Channel (samples) |
| $4014 | OAMDMA | Sprite DMA (writes 256 bytes to OAM) |
| $4015 | APU Status | Enable/disable audio channels |
| $4016 | Controller 1 | Controller 1 input |
| $4017 | Controller 2 / Frame Counter | Controller 2 and APU frame counter |

## Cartridge Memory ($6000-$FFFF)

### SRAM ($6000-$7FFF)
- Optional 8KB of battery-backed RAM
- Used for save data in games like Zelda and Final Fantasy
- Not all cartridges have this

### PRG-ROM ($8000-$FFFF)
- 32KB of program code and data from the cartridge
- May be bank-switched by mapper to access more than 32KB
- Contains the game code, graphics data, and sound data

**Fixed Vectors** (always at these addresses):
- $FFFA-$FFFB: NMI vector (called during VBlank)
- $FFFC-$FFFD: Reset vector (program start)
- $FFFE-$FFFF: IRQ/BRK vector (interrupt handler)

## PPU Memory Map (16KB Address Space)

The PPU has its own separate address space, accessed through CPU registers:

```
Address Range    Size     Description
--------------------------------------------------------------------------------
$0000-$0FFF      4KB      Pattern Table 0 (sprite/bg graphics)
$1000-$1FFF      4KB      Pattern Table 1 (sprite/bg graphics)

$2000-$23FF      1KB      Nametable 0 (screen layout)
$2400-$27FF      1KB      Nametable 1
$2800-$2BFF      1KB      Nametable 2
$2C00-$2FFF      1KB      Nametable 3

$3000-$3EFF      -        Mirror of $2000-$2EFF
$3F00-$3F1F      32B      Palette RAM (sprite and background colors)
$3F20-$3FFF      -        Mirrors of $3F00-$3F1F
```

### Pattern Tables ($0000-$1FFF)
- Stores 8x8 pixel tile graphics
- Usually from CHR-ROM on cartridge (or CHR-RAM)
- Each tile is 16 bytes (8 bytes for low bit plane, 8 for high)

### Nametables ($2000-$2FFF)
- Define which tiles appear on screen
- Each nametable represents one screen (256x240 pixels)
- Only 2KB of physical RAM, others are mirrored (or on cartridge)

**Mirroring Modes**:
- **Horizontal**: Nametables 0 and 1 are the same, 2 and 3 are the same
- **Vertical**: Nametables 0 and 2 are the same, 1 and 3 are the same
- **Single Screen**: All nametables mirror the same memory
- **Four Screen**: All four nametables are unique (requires extra RAM)

### Palette RAM ($3F00-$3F1F)
- 32 bytes defining colors for sprites and background
- $3F00-$3F0F: Background palettes (4 palettes, 4 colors each)
- $3F10-$3F1F: Sprite palettes (4 palettes, 4 colors each)
- First color of each palette ($3F00, $3F04, etc.) is shared as backdrop

**Special Note**: Addresses $3F04, $3F08, $3F0C mirror to $3F00 (backdrop color).

## Memory Access Patterns

### CPU to PPU Communication
1. Write address to $2006 (twice: high byte, then low byte)
2. Read/write data through $2007
3. Address auto-increments after each access (by 1 or 32, configurable)

### DMA Transfer ($4014)
- Fast way to copy 256 bytes to sprite memory (OAM)
- Write high byte of source address to $4014
- CPU is stalled for 513-514 cycles during transfer
- Commonly used: Write $02 to copy $0200-$02FF to OAM

## Bank Switching (Mappers)

Many games need more than 32KB of ROM. Mappers allow bank switching:
- **Mapper 0 (NROM)**: No banking, simple 16KB or 32KB ROM
- **Mapper 1 (MMC1)**: Switchable 16KB ROM banks
- **Mapper 2 (UxROM)**: Switchable 16KB ROM bank + fixed 16KB
- **Mapper 4 (MMC3)**: Advanced banking with IRQ support
