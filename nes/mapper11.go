package nes

// Mapper11 implements the Color Dreams mapper.
// This mapper is used by several unlicensed NES games including:
//   - Crystal Mines
//   - Baby Boomer
//   - Menace Beach
//   - Raid 2020
//   - Robodemons
//   - Metal Fighter
//
// Color Dreams features:
//   - Switchable 32KB PRG ROM banks
//   - Switchable 8KB CHR ROM banks
//   - No mirroring control (determined by cartridge hardware)
//   - Very simple: single register controls both PRG and CHR
//
// Memory Map:
//
//	$0000-$1FFF: 8KB switchable CHR ROM bank
//	$6000-$7FFF: 8KB PRG RAM (if present, rare)
//	$8000-$FFFF: 32KB switchable PRG ROM bank
//
// Register ($8000-$FFFF, write):
//
//	Bits 0-1: CHR bank select (selects 8KB bank)
//	Bits 4-5: PRG bank select (selects 32KB bank)
//	Other bits: Unused
type Mapper11 struct {
	*Cartridge
	prgBank  int // Currently selected PRG bank (0-3 typical)
	chrBank  int // Currently selected CHR bank (0-3 typical)
	prgBanks int // Total number of 32KB PRG banks available
	chrBanks int // Total number of 8KB CHR banks available
}

// NewMapper11 creates a new Color Dreams mapper instance.
func NewMapper11(cartridge *Cartridge) Mapper {
	prgBanks := len(cartridge.PRG) / 0x8000 // Calculate number of 32KB banks
	if prgBanks == 0 {
		prgBanks = 1
	}

	chrBanks := len(cartridge.CHR) / 0x2000 // Calculate number of 8KB banks
	if chrBanks == 0 {
		chrBanks = 1
	}

	return &Mapper11{
		Cartridge: cartridge,
		prgBank:   0,
		chrBank:   0,
		prgBanks:  prgBanks,
		chrBanks:  chrBanks,
	}
}

// Read handles memory reads from the cartridge address space.
func (m *Mapper11) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		// CHR ROM: Read from currently selected 8KB bank
		index := m.chrBank*0x2000 + int(address)
		return m.CHR[index]
	case address >= 0x8000:
		// PRG ROM: Read from currently selected 32KB bank
		index := m.prgBank*0x8000 + int(address-0x8000)
		return m.PRG[index]
	case address >= 0x6000:
		// PRG RAM (if present, though rare)
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		panicUnhandledAddress("mapper 11 read", address)
	}
	return 0
}

// Write handles memory writes to the cartridge address space.
// Writes to $8000-$FFFF control both PRG and CHR banking.
func (m *Mapper11) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// CHR ROM: Write to currently selected bank (if CHR RAM)
		index := m.chrBank*0x2000 + int(address)
		m.CHR[index] = value
	case address >= 0x8000:
		// Bank select register
		// Bits 0-1: Select 8KB CHR bank
		m.chrBank = int(value&0x03) % m.chrBanks

		// Bits 4-5: Select 32KB PRG bank
		m.prgBank = int((value>>4)&0x03) % m.prgBanks
	case address >= 0x6000:
		// PRG RAM (if present)
		index := int(address) - 0x6000
		m.SRAM[index] = value
	default:
		panicUnhandledAddress("mapper 11 write", address)
	}
}
