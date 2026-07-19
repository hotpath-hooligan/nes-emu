package nes

// CNROM features:
//   - Fixed 16KB or 32KB PRG ROM (no bank switching)
//   - Switchable 8KB CHR ROM banks (2, 4, or 8 banks typical)
//   - No mirroring control (determined by cartridge hardware)
//   - Very simple: writes to $8000-$FFFF select CHR bank
//
// Memory Map:
//
//	$0000-$1FFF: 8KB switchable CHR ROM bank
//	$6000-$7FFF: 8KB PRG RAM (if present)
//	$8000-$FFFF: 16KB or 32KB fixed PRG ROM
type Mapper3 struct {
	*Cartridge
	chrBank  int // Currently selected CHR bank (0-255)
	chrBanks int // Total number of 8KB CHR banks available
}

// NewMapper3 creates a new CNROM mapper instance.
func NewMapper3(cartridge *Cartridge) Mapper {
	chrBanks := len(cartridge.CHR) / 0x2000 // Calculate number of 8KB banks
	if chrBanks == 0 {
		chrBanks = 1 // Use CHR RAM if no CHR ROM
	}

	return &Mapper3{
		Cartridge: cartridge,
		chrBank:   0,
		chrBanks:  chrBanks,
	}
}

// Read handles memory reads from the cartridge address space.
// CHR reads come from the selected bank, PRG reads from fixed ROM.
func (m *Mapper3) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		// CHR ROM: Read from currently selected 8KB bank
		bankOffset := m.chrBank * 0x2000
		return m.CHR[bankOffset+int(address)]
	case address >= 0x8000:
		// PRG ROM: Fixed, no banking
		// Support both 16KB and 32KB PRG ROM
		index := int(address - 0x8000)
		if len(m.PRG) <= 0x4000 {
			// 16KB PRG ROM: mirror it to fill $8000-$FFFF
			index %= 0x4000
		}
		return m.PRG[index]
	case address >= 0x6000:
		// PRG RAM (battery-backed SRAM)
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		panicUnhandledAddress("mapper 3 read", address)
	}
	return 0
}

// Write handles memory writes to the cartridge address space.
// Writes to $8000-$FFFF select the CHR bank (only lower bits used).
func (m *Mapper3) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// CHR ROM: Write to currently selected bank (if CHR RAM)
		bankOffset := m.chrBank * 0x2000
		m.CHR[bankOffset+int(address)] = value
	case address >= 0x8000:
		// Bank select: Write to $8000-$FFFF selects CHR bank
		// Only use the bits needed (typically 2 bits for 4 banks)
		m.chrBank = int(value) % m.chrBanks
	case address >= 0x6000:
		// PRG RAM (battery-backed SRAM)
		index := int(address) - 0x6000
		m.SRAM[index] = value
	default:
		panicUnhandledAddress("mapper 3 write", address)
	}
}
