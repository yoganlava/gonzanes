package emulator

const (
	CHRRAMStart   = 0x0000
	CHRRAMEnd     = 0x1FFF
	VRAMStart     = 0x2000
	VRAMMirrorsEnd = 0x3EFF
	PaletteStart  = 0x3F20
	PaletteMirrorsEnd    = 0x3FFF
)

type PPU struct {
	// chr ram
	chrRAM []uint8
	// internal vram
	vram [2048]uint8
	// internal palette control
	paletteRAM [32]uint8
	// Object Attribute Memory
	oam [256]uint8

	// Registers
	// PPUCTRL
	ppuCtrl uint8
	// PPUMASK
	ppuMask uint8
	// PPUSTATUS
	ppuStatus uint8
	// OAMADDR
	oamAddr uint8
	// PPUSCROLL
	ppuScrollX uint8
	ppuScrollY uint8
	// Scroll latch reset during PPUSTATUS read
	ppuScrollLatch bool
	// PPUADDR
	// high byte, low byte
	ppuAddr     uint16
	ppuAddrHigh bool
}