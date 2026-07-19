package nes

import "fmt"

// Mapper represents a memory mapper (MMC - Memory Management Controller) for NES cartridges.
// Different mappers provide different memory banking capabilities, allowing games to use
// more PRG ROM (program code) or CHR ROM (graphics data) than the NES address space allows.
//
// The NES has a 16-bit address space, but mappers enable bank switching to access
// larger ROM sizes. Each mapper type has unique behavior and capabilities.
type Mapper interface {
	// Read returns the byte at the specified address, handling bank switching as needed
	Read(address uint16) byte

	// Write performs a write to the specified address, often triggering bank switches
	Write(address uint16, value byte)
}

func panicUnhandledAddress(component string, address uint16) {
	panic(fmt.Sprintf("%s: unhandled address 0x%04X", component, address))
}

// NewMapper creates the appropriate mapper implementation based on the ROM's mapper number.
// Currently supported mappers:
//   - 0: NROM (no mapper, simple 16KB or 32KB PRG) - Super Mario Bros, Donkey Kong
//   - 1: MMC1 (used by games like Zelda, Metroid, Final Fantasy)
//   - 2: UxROM (used by games like Mega Man, Castlevania, Contra)
//   - 3: CNROM (used by games like Gradius, Solomon's Key, Paperboy, Arkanoid)
//   - 7: AxROM (used by games like Battletoads, Wizards & Warriors, Marble Madness)
//   - 11: Color Dreams (used by games like Crystal Mines, Baby Boomer, Menace Beach)
func NewMapper(sys *System) (Mapper, error) {
	rom := sys.Cartridge
	switch rom.Mapper {
	case 0:
		return NewMapper2(rom), nil // NROM uses same implementation as UxROM
	case 1:
		return NewMapper1(rom), nil
	case 2:
		return NewMapper2(rom), nil
	case 3:
		return NewMapper3(rom), nil
	case 7:
		return NewMapper7(rom), nil
	case 11:
		return NewMapper11(rom), nil
	}
	err := fmt.Errorf("unsupported mapper: %d (only mappers 0, 1, 2, 3, 7, and 11 are supported)", rom.Mapper)
	return nil, err
}
