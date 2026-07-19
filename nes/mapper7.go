package nes

// Mapper7 implements the AxROM mapper (also known as AOROM).
// This mapper is used by several popular NES games including:
//   - Battletoads
//   - Wizards & Warriors
//   - Marble Madness
//   - Solstice
//   - Time Lord
//   - Arch Rivals
//
// AxROM features:
//   - Switchable 32KB PRG ROM banks
//   - No CHR ROM banking (uses 8KB CHR RAM)
//   - Single-screen mirroring control
//   - Very simple: single register controls everything
//
// Memory Map:
//
//	$0000-$1FFF: 8KB CHR RAM (fixed, no banking)
//	$6000-$7FFF: 8KB PRG RAM (if present, rare for AxROM)
//	$8000-$FFFF: 32KB switchable PRG ROM bank
//
// Register ($8000-$FFFF, write):
//
//	Bits 0-2: PRG bank select (selects 32KB bank)
//	Bit 4: Single-screen mirroring (0 = nametable A, 1 = nametable B)
//	Other bits: Unused
type Mapper7 struct {
	*Cartridge
	prgBank    int  // Currently selected PRG bank (0-7 typical)
	prgBanks   int  // Total number of 32KB PRG banks available
	mirrorMode byte // Single-screen mirror: 0 or 1
}

// NewMapper7 creates a new AxROM mapper instance.
func NewMapper7(cartridge *Cartridge) Mapper {
	prgBanks := len(cartridge.PRG) / 0x8000 // Calculate number of 32KB banks
	if prgBanks == 0 {
		prgBanks = 1
	}

	return &Mapper7{
		Cartridge:  cartridge,
		prgBank:    0,
		prgBanks:   prgBanks,
		mirrorMode: 0,
	}
}

// Read handles memory reads from the cartridge address space.
func (m *Mapper7) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		// CHR RAM: Fixed 8KB, no banking
		return m.CHR[address]
	case address >= 0x8000:
		// PRG ROM: Read from currently selected 32KB bank
		index := m.prgBank*0x8000 + int(address-0x8000)
		return m.PRG[index]
	case address >= 0x6000:
		// PRG RAM (if present, though rare for AxROM carts)
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		panicUnhandledAddress("mapper 7 read", address)
	}
	return 0
}

// Write handles memory writes to the cartridge address space.
// Writes to $8000-$FFFF control PRG banking and mirroring.
func (m *Mapper7) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// CHR RAM: Writable
		m.CHR[address] = value
	case address >= 0x8000:
		// Bank select + mirroring control
		// Bits 0-2: Select 32KB PRG bank
		m.prgBank = int(value&0x07) % m.prgBanks

		// Bit 4: Single-screen mirroring control
		// 0 = use nametable A ($2000), 1 = use nametable B ($2400)
		m.mirrorMode = (value >> 4) & 0x01

		// Update the cartridge mirror mode for single-screen mirroring
		if m.mirrorMode == 0 {
			m.Cartridge.Mirror = MirrorSingle0
		} else {
			m.Cartridge.Mirror = MirrorSingle1
		}
	case address >= 0x6000:
		// PRG RAM (if present)
		index := int(address) - 0x6000
		m.SRAM[index] = value
	default:
		panicUnhandledAddress("mapper 7 write", address)
	}
}
