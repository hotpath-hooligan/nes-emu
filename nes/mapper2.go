package nes

// UxROM features:
//   - Switchable 16KB PRG ROM bank at $8000-$BFFF
//   - Fixed 16KB PRG ROM bank at $C000-$FFFF (last bank)
//   - 8KB CHR RAM (not banked)
//   - No mirroring control (determined by cartridge hardware)
//
// Memory Map:
//
//	$6000-$7FFF: 8KB PRG RAM (battery backed SRAM)
//	$8000-$BFFF: 16KB switchable PRG ROM bank
//	$C000-$FFFF: 16KB fixed PRG ROM bank (last bank)
type Mapper2 struct {
	*Cartridge
	prgBanks int // Total number of 16KB PRG banks
	prgBank1 int // Currently selected bank for $8000-$BFFF
	prgBank2 int // Fixed bank for $C000-$FFFF (always last bank)
}

func NewMapper2(cartridge *Cartridge) Mapper {
	prgBanks := len(cartridge.PRG) / 0x4000 // Calculate number of 16KB banks
	prgBank1 := 0                           // First bank starts at 0
	prgBank2 := prgBanks - 1                // Last bank fixed at end
	return &Mapper2{cartridge, prgBanks, prgBank1, prgBank2}
}

// switchPRGBank changes the switchable PRG ROM bank.
func (m *Mapper2) switchPRGBank(bank int) {
	m.prgBank1 = bank % m.prgBanks
}

// Read handles memory reads from the cartridge address space.
// Routes reads to CHR, PRG ROM banks, or SRAM based on address.
func (m *Mapper2) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return m.CHR[address]
	case address >= 0xC000:
		index := m.prgBank2*0x4000 + int(address-0xC000)
		return m.PRG[index]
	case address >= 0x8000:
		index := m.prgBank1*0x4000 + int(address-0x8000)
		return m.PRG[index]
	case address >= 0x6000:
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		panicUnhandledAddress("mapper 2 read", address)
	}
	return 0
}

// Write handles memory writes to the cartridge address space.
// Writes to $8000-$FFFF select the PRG bank, other writes go to CHR RAM or SRAM.
func (m *Mapper2) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		m.CHR[address] = value
	case address >= 0x8000:
		m.switchPRGBank(int(value))
	case address >= 0x6000:
		index := int(address) - 0x6000
		m.SRAM[index] = value
	default:
		panicUnhandledAddress("mapper 2 write", address)
	}
}
