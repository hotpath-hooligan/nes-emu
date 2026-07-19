package nes

// MMC1 features:
//   - PRG ROM bank switching (16KB or 32KB modes)
//   - CHR ROM bank switching (4KB or 8KB modes)
//   - Configurable mirroring (horizontal, vertical, or single-screen)
//   - Serial port for configuration (writes to control registers via shift register)
type Mapper1 struct {
	*Cartridge
	shiftRegister byte   // 5-bit shift register for loading control values
	control       byte   // Control register (mirroring and bank modes)
	prgMode       byte   // PRG ROM bank mode (0-3)
	chrMode       byte   // CHR ROM bank mode (0-1)
	prgBank       byte   // Selected PRG ROM bank
	chrBank0      byte   // First CHR bank (or both banks in 8KB mode)
	chrBank1      byte   // Second CHR bank (only used in 4KB mode)
	prgOffsets    [2]int // Calculated offsets for the two 16KB PRG banks
	chrOffsets    [2]int // Calculated offsets for the two 4KB CHR banks
}

func NewMapper1(cartridge *Cartridge) Mapper {
	m := Mapper1{}
	m.Cartridge = cartridge
	m.shiftRegister = 0x10                // Initialize with bit 5 set
	m.prgOffsets[1] = m.prgBankOffset(-1) // Last bank fixed in reset state
	return &m
}

// prgBankOffset calculates the byte offset for a PRG ROM bank.
// Handles negative indices (e.g., -1 = last bank) and wrapping.
func (m *Mapper1) prgBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.PRG) / 0x4000
	offset := index * 0x4000
	if offset < 0 {
		offset += len(m.PRG)
	}
	return offset
}

// chrBankOffset calculates the byte offset for a CHR ROM bank.
// Handles negative indices and wrapping similar to PRG banks.
func (m *Mapper1) chrBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.CHR) / 0x1000
	offset := index * 0x1000
	if offset < 0 {
		offset += len(m.CHR)
	}
	return offset
}

// updateOffsets recalculates PRG and CHR bank offsets based on current mode and bank values.
// Called after any write to control registers.
func (m *Mapper1) updateOffsets() {
	switch m.prgMode {
	case 0, 1:
		m.prgOffsets[0] = m.prgBankOffset(int(m.prgBank & 0xFE))
		m.prgOffsets[1] = m.prgBankOffset(int(m.prgBank | 0x01))
	case 2:
		m.prgOffsets[0] = 0
		m.prgOffsets[1] = m.prgBankOffset(int(m.prgBank))
	case 3:
		m.prgOffsets[0] = m.prgBankOffset(int(m.prgBank))
		m.prgOffsets[1] = m.prgBankOffset(-1)
	}
	switch m.chrMode {
	case 0:
		m.chrOffsets[0] = m.chrBankOffset(int(m.chrBank0 & 0xFE))
		m.chrOffsets[1] = m.chrBankOffset(int(m.chrBank0 | 0x01))
	case 1:
		m.chrOffsets[0] = m.chrBankOffset(int(m.chrBank0))
		m.chrOffsets[1] = m.chrBankOffset(int(m.chrBank1))
	}
}

func (m *Mapper1) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		bank := address / 0x1000
		offset := address % 0x1000
		return m.CHR[m.chrOffsets[bank]+int(offset)]
	case address >= 0x8000:
		address = address - 0x8000
		bank := address / 0x4000
		offset := address % 0x4000
		return m.PRG[m.prgOffsets[bank]+int(offset)]
	case address >= 0x6000:
		return m.SRAM[int(address)-0x6000]
	default:
		panicUnhandledAddress("mapper 1 read", address)
	}
	return 0
}

func (m *Mapper1) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		bank := address / 0x1000
		offset := address % 0x1000
		m.CHR[m.chrOffsets[bank]+int(offset)] = value
	case address >= 0x8000:
		m.loadRegister(address, value)
	case address >= 0x6000:
		m.SRAM[int(address)-0x6000] = value
	default:
		panicUnhandledAddress("mapper 1 write", address)
	}
}

// loadRegister handles the serial write protocol for MMC1.
// The MMC1 uses a 5-bit shift register that requires 5 consecutive writes
// to load a value into a control register. Bit 7 of the value resets the register.
func (m *Mapper1) loadRegister(address uint16, value byte) {
	if value&0x80 == 0x80 {
		// Reset: bit 7 set clears shift register and sets control to mode 3
		m.shiftRegister = 0x10
		m.writeControl(m.control | 0x0C)
	} else {
		// Shift in one bit
		complete := m.shiftRegister&1 == 1
		m.shiftRegister >>= 1
		m.shiftRegister |= (value & 1) << 4
		if complete {
			// All 5 bits received, write to the selected register
			m.writeRegister(address, m.shiftRegister)
			m.shiftRegister = 0x10
		}
	}
}

// writeRegister routes the completed shift register value to the appropriate control register
// based on the address range.
func (m *Mapper1) writeRegister(address uint16, value byte) {
	switch {
	case address <= 0x9FFF:
		m.writeControl(value)
	case address <= 0xBFFF:
		m.writeCHRBank0(value)
	case address <= 0xDFFF:
		m.writeCHRBank1(value)
	case address <= 0xFFFF:
		m.writePRGBank(value)
	}
}

func (m *Mapper1) writeControl(value byte) {
	m.control = value
	m.chrMode = (value >> 4) & 1
	m.prgMode = (value >> 2) & 3
	mirror := value & 3
	switch mirror {
	case 0:
		m.Mirror = MirrorSingle0
	case 1:
		m.Mirror = MirrorSingle1
	case 2:
		m.Mirror = MirrorVertical
	case 3:
		m.Mirror = MirrorHorizontal
	}
	m.updateOffsets()
}

func (m *Mapper1) writeCHRBank0(value byte) {
	m.chrBank0 = value
	m.updateOffsets()
}

func (m *Mapper1) writeCHRBank1(value byte) {
	m.chrBank1 = value
	m.updateOffsets()
}

func (m *Mapper1) writePRGBank(value byte) {
	m.prgBank = value & 0x0F
	m.updateOffsets()
}
