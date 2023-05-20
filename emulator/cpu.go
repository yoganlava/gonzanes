package emulator

import "fmt"

type AddressingMode int
type StatusFlag int


const (
	Immediate AddressingMode = iota
	ZeroPage
	ZeroPageX
	ZeroPageY
	Relative
	Absolute
	AbsoluteX
	AbsoluteY
	IndirectX
	InderectY
)

const (
	StatusNegativeFlag = 0b10000000
	StatusOverflowFlag = 0b01000000
	StatusBreak = 0b00010000
	StatusDecimalMode = 0b00001000
	StatusInterruptDisable = 0b00000100
	StatusZeroFlag = 0b00000010
	StatusCarryFlag = 0b00000001
)

type CPU struct {
	bus Bus
	stack [256]uint8

	// accumulator
	ac uint8
	// index x
	ix uint8
	// index y
	iy uint8
	// two-byte program counter
	pc uint16
	// stack pointer
	sp uint16
	// status register
	// 1 bit per status
	// carry flag, zero flag, interrupt disable, decimal mode, brk command, overflow flag, neg flag
	sr uint8
}

func Init() *CPU {
	return &CPU{}
}

func (c *CPU) Reset() {
	c.ac = 0
	c.ix = 0
	c.iy = 0
	c.sr = 0
	c.sp = 0

	// reset pc to addr stored in reset vector
	c.pc = c.bus.readU16(0xFFFC)
}

func (c *CPU) fetch() uint8 {
	byte := c.bus.read(c.pc)
	c.pc++
	return byte
}

func (c *CPU) stackPush(byte uint8) {
	c.stack[c.sp] = byte
	c.sp++
}

func (c *CPU) stackPop() uint8 {
	c.sp--
	return c.stack[c.sp]
}

func (c *CPU) stackPushU16(value uint16) {
	hi := uint8(value >> 8)
	lo := uint8(value & 0xff)
	c.stackPush(hi)
	c.stackPush(lo)
}

func (c *CPU) stackPopU16() uint16 {
	lo := c.stackPop()
	hi := c.stackPop()
	return uint16(hi << 8) | uint16(lo)
}


func (c *CPU) LoadRom(rom *ROM) {
	c.bus.LoadRom(rom.prgROM)
	c.pc = c.bus.readU16(0xFFFC)

	fmt.Printf("Reset Vector: %X\n", c.pc)
}

func (c *CPU) getOperandAddressInMode(addressingMode AddressingMode) uint16 {
	switch addressingMode {
		case Immediate:
			return c.pc
		case Absolute:
			return c.bus.readU16(c.pc)
	}

	panic("Invalid addressing mode")
}

// Set zero and negative flag if needed
func (c *CPU) setZNStatus(value uint8) {
	if value == 0 {
		c.sr |= StatusZeroFlag
	}
	if (value >> 7) == 0x1 {
		c.sr |= StatusNegativeFlag
	}
}

func (c *CPU) isSRFlagSet(flag uint8) bool {
	return (flag & c.sr) != 0
}

func (c *CPU) lda(addressingMode AddressingMode) {
	c.ac = c.bus.read(c.getOperandAddressInMode(addressingMode))
	c.setZNStatus(c.ac)
	c.pc += 1
}

func (c *CPU) sta(addressingMode AddressingMode) {
	c.bus.Write(c.bus.readU16(c.getOperandAddressInMode(addressingMode)), c.ac)
	c.pc += 2
}


func (c *CPU) Step() {
	// fetch opcode
	currentOpcode := c.fetch()
	fmt.Printf("Current OpCode: %X\n", currentOpcode)
	switch currentOpcode {
		// LDA Immediate
		case 0xA9:
			c.lda(Immediate)
		// SEI
		// Set the interrupt disable flag to one.
		case 0x78:
			c.sr |= StatusInterruptDisable
		// JMP Absolute
		case 0x4C:
			c.pc = c.bus.readU16(c.pc)
		// STA Absolute
		case 0x8D:
			c.sta(Absolute)
		// BNE
		case 0xD8:
			offset := c.bus.readU16(c.pc)
			if c.isSRFlagSet(StatusZeroFlag) {
				c.pc += offset
			}
		// JSR
		case 0x20:
			// address - 1
			c.stackPushU16(c.pc + 2 - 1)
			c.pc = c.bus.readU16(c.pc)
		default:
			panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpcode))
	}
}