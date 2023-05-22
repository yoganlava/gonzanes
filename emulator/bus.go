package emulator

import "fmt"

const (
	RamStart          = 0x0000
	RamEnd            = 0x07FF
	RamMirrorsEnd     = 0x1FFF
	PPURegistersStart = 0x2000
	PPURegistersEnd   = 0x3FFF
	PRGROMStart       = 0x8000
	PRGROMEnd         = 0xFFFF
	PRGRAMStart       = 0x6000
	PRGRAMEnd         = 0x7FFF
	APUIOStart        = 0x4000
	APUIOEnd          = 0x4018
)

type Bus struct {
	cpuRAM [2048]uint8
	prgROM []uint8
	prgRAM [8192]uint8
	ppu    PPU
}

func (b *Bus) LoadRom(rom *ROM) {
	b.prgROM = rom.prgROM
	b.ppu.chrRAM = rom.chrROM
}

func (b *Bus) Read(addr uint16) uint8 {
	if addr >= RamStart && addr <= RamMirrorsEnd {
		return b.cpuRAM[addr%RamEnd]
	} else if addr >= PPURegistersStart && addr <= PPURegistersEnd {
		return b.ReadPPURegister((addr - PPURegistersStart) % 8)
	} else if addr >= PRGRAMStart && addr <= PRGRAMEnd {
		return b.prgRAM[addr-PRGRAMStart]
	} else if addr >= APUIOStart && addr <= APUIOEnd {
		fmt.Println("APU/IO NOT SUPPORTED YET")
		return 0
	} else if addr >= PRGROMStart && addr <= PRGROMEnd {
		return b.prgROM[addr-PRGROMStart]
	} else {
		panic(fmt.Sprintf("Unsupported Read Addr: %X", addr))
	}
}

func (b *Bus) Write(addr uint16, data uint8) {
	if addr >= RamStart && addr <= RamMirrorsEnd {
		b.cpuRAM[addr] = data
	} else if addr >= PPURegistersStart && addr <= PPURegistersEnd {
		b.WritePPURegister((addr-PPURegistersStart)%8, data)
	} else if addr >= PRGRAMStart && addr <= PRGRAMEnd {
		b.prgRAM[addr-PRGRAMStart] = data
	} else if addr >= APUIOStart && addr <= APUIOEnd {
		fmt.Println("APU/IO NOT SUPPORTED YET")
	} else if addr >= PRGROMStart && addr <= PRGROMEnd {
		panic(fmt.Sprintf("Cannot write into PRG ROM"))
	} else {
		panic(fmt.Sprintf("Unsupported Write Addr: %X", addr))
	}

}

// TODO mirror based on mirroring type
func mirrorAddr(addr uint16) uint16 {
	return addr
}

func (b *Bus) ReadPPURegister(index uint16) uint8 {
	switch index {
	case 0, 1, 3, 5, 6, 8:
		return 0
	case 2:
		// TODO read resets write pair for $2005/$2006
		b.ppu.ppuScrollLatch = false
		return b.ppu.ppuStatus
	case 4:
		return b.ppu.oam[b.ppu.oamAddr]
	case 7:
		if b.ppu.ppuAddr >= CHRRAMStart && b.ppu.ppuAddr <= CHRRAMEnd {
			return b.ppu.chrRAM[b.ppu.ppuAddr]
		} else if b.ppu.ppuAddr >= VRAMStart && b.ppu.ppuAddr <= VRAMMirrorsEnd {
			return b.ppu.vram[mirrorAddr((b.ppu.ppuAddr - VRAMStart) % 2048)]
		} else if b.ppu.ppuAddr >= PaletteStart && b.ppu.ppuAddr <= PaletteMirrorsEnd {
			return b.ppu.paletteRAM[(b.ppu.ppuAddr - PaletteStart) % 32]
		}
	}
	panic("Unreachable")
}

func (b *Bus) WritePPURegister(index uint16, data uint8) {
	switch index {
	case 0:
		// TODO NMI
		b.ppu.ppuCtrl = data
	case 1:
		b.ppu.ppuMask = data
	case 2:
		panic("PPUSTATUS is readonly")
	case 3:
		b.ppu.oamAddr = data
	case 4:
		b.ppu.oam[b.ppu.oamAddr] = data
		b.ppu.oamAddr++
	case 5:
		if b.ppu.ppuScrollLatch {
			b.ppu.ppuScrollY = data
		} else {
			b.ppu.ppuScrollX |= data
		}

		b.ppu.ppuScrollLatch = !b.ppu.ppuScrollLatch
	case 6:
		if b.ppu.ppuAddr >= CHRRAMStart && b.ppu.ppuAddr <= CHRRAMEnd {
			panic("Cannot write to CHRRAM")
		} else if b.ppu.ppuAddr >= VRAMStart && b.ppu.ppuAddr <= VRAMMirrorsEnd {
			b.ppu.vram[mirrorAddr((b.ppu.ppuAddr - VRAMStart) % 2048)] = data
		} else if b.ppu.ppuAddr >= PaletteStart && b.ppu.ppuAddr <= PaletteMirrorsEnd {
			b.ppu.paletteRAM[(b.ppu.ppuAddr - PaletteStart) % 32] = data
		}
	case 7:
		if b.ppu.ppuAddrHigh {
			b.ppu.ppuAddr |= uint16(data) << 8
		} else {
			b.ppu.ppuAddr |= uint16(data)
		}

		b.ppu.ppuAddrHigh = !b.ppu.ppuAddrHigh
	// OAMDMA
	case 8:
		hi := uint16(data) << 8
		for i := uint16(0); i < 256; i++ {
			b.ppu.oam[i] = b.Read(hi + i)
		}
	}

	panic("Unreachable")
}


func (b *Bus) ReadU16(addr uint16) uint16 {
	return uint16(b.Read(addr+1))<<8 | uint16(b.Read(addr))
}

func (b *Bus) WriteU16(addr uint16, data uint16) {
	hi := uint8(data >> 8)
	lo := uint8(data & 0xff)
	b.Write(addr, lo)
	b.Write(addr+1, hi)
}
