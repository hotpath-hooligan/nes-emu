package nes

type Memory interface {
	Read(address uint16) byte
	Write(address uint16, value byte)
}

type cpuMemory struct {
	system *System
}

func NewCPUMemory(sys *System) Memory {
	return &cpuMemory{sys}
}

func (mem *cpuMemory) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return mem.system.RAM[address%0x0800]
	case address < 0x4000:
		return mem.system.PPU.readRegister(0x2000 + address%8)
	case address == 0x4014:
		return mem.system.PPU.readRegister(address)
	case address == 0x4015:
		return mem.system.APU.readRegister(address)
	case address == 0x4016:
		return mem.system.Controller1.Read()
	case address == 0x4017:
		return mem.system.Controller2.Read()
	case address < 0x6000:
	case address >= 0x6000:
		return mem.system.Mapper.Read(address)
	default:
		panicUnhandledAddress("CPU memory read", address)
	}
	return 0
}

func (mem *cpuMemory) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		mem.system.RAM[address%0x0800] = value
	case address < 0x4000:
		mem.system.PPU.writeRegister(0x2000+address%8, value)
	case address < 0x4014:
		mem.system.APU.writeRegister(address, value)
	case address == 0x4014:
		mem.system.PPU.writeRegister(address, value)
	case address == 0x4015:
		mem.system.APU.writeRegister(address, value)
	case address == 0x4016:
		mem.system.Controller1.Write(value)
		mem.system.Controller2.Write(value)
	case address == 0x4017:
		mem.system.APU.writeRegister(address, value)
	case address < 0x6000:
	case address >= 0x6000:
		mem.system.Mapper.Write(address, value)
	default:
		panicUnhandledAddress("CPU memory write", address)
	}
}

type ppuMemory struct {
	system *System
}

func NewPPUMemory(sys *System) Memory {
	return &ppuMemory{sys}
}

func (mem *ppuMemory) Read(address uint16) byte {
	address = address % 0x4000
	switch {
	case address < 0x2000:
		return mem.system.Mapper.Read(address)
	case address < 0x3F00:
		mode := mem.system.Cartridge.Mirror
		return mem.system.PPU.nameTableData[MirrorAddress(mode, address)%2048]
	case address < 0x4000:
		return mem.system.PPU.readPalette(address % 32)
	default:
		panicUnhandledAddress("PPU memory read", address)
	}
	return 0
}

func (mem *ppuMemory) Write(address uint16, value byte) {
	address = address % 0x4000
	switch {
	case address < 0x2000:
		mem.system.Mapper.Write(address, value)
	case address < 0x3F00:
		mode := mem.system.Cartridge.Mirror
		mem.system.PPU.nameTableData[MirrorAddress(mode, address)%2048] = value
	case address < 0x4000:
		mem.system.PPU.writePalette(address%32, value)
	default:
		panicUnhandledAddress("PPU memory write", address)
	}
}

const (
	MirrorHorizontal = 0
	MirrorVertical   = 1
	MirrorSingle0    = 2
	MirrorSingle1    = 3
	MirrorFour       = 4
)

var MirrorLookup = [...][4]uint16{
	{0, 0, 1, 1},
	{0, 1, 0, 1},
	{0, 0, 0, 0},
	{1, 1, 1, 1},
	{0, 1, 2, 3},
}

func MirrorAddress(mode byte, address uint16) uint16 {
	address = (address - 0x2000) % 0x1000
	table := address / 0x0400
	offset := address % 0x0400
	return 0x2000 + MirrorLookup[mode][table]*0x0400 + offset
}
