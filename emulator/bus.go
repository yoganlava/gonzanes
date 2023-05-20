package emulator

import "fmt"

const (
	RamStart = 0x0000
	RamEnd = 0x7FFF
	RamMirrorsEnd = 0x17FF
	PPURegistersStart = 0x2000
	PPURegistersEnd = 0x3FFF
	// CartridgeSpaceStart = 0x4020
	PRGROMStart = 0x8000
	PRGROMEnd = 0xFFFF
	// CartridgeSpaceEnd = 0xFFFF
)

type Bus struct {
	cpuRAM [2048]uint8
	prgROM []uint8
}

func (b *Bus) LoadRom(prgROM []uint8) {
	b.prgROM = prgROM
}

func (b *Bus) read(addr uint16) uint8 {
	if addr >= RamStart && addr <= RamMirrorsEnd {
		return b.cpuRAM[addr % RamEnd]
	} else if addr >= PPURegistersStart && addr <= PPURegistersEnd {
		fmt.Println("PPU READ NOT SUPPORTED YET")
		return 0
	} else if addr >= PRGROMStart && addr <= PRGROMEnd {
		return b.prgROM[addr - PRGROMStart]
	} else {
		panic(fmt.Sprintf("Unsupported Read Addr: %X", addr))
	}
}

func (b *Bus) Write(addr uint16, data uint8) {
	if addr >= RamStart && addr <= RamMirrorsEnd {
		b.cpuRAM[addr] = data
	} else if addr >= PPURegistersStart && addr <= PPURegistersEnd {
		fmt.Println("PPU WRITE NOT SUPPORTED YET")
	} else if addr >= PRGROMStart && addr <= PRGROMEnd {
		panic(fmt.Sprintf("Cannot write into PRG ROM"))
	} else {
		panic(fmt.Sprintf("Unsupported Write Addr: %X", addr))
	}
	
}

func (b *Bus) readU16(addr uint16) uint16 {
	return uint16(b.read(addr + 1)) << 8 | uint16(b.read(addr))
}

func (b *Bus) writeU16(addr uint16, data uint16) {
	hi := uint8(data >> 8);
	lo := uint8(data & 0xff);
	b.Write(addr, lo)
	b.Write(addr + 1, hi)
}
