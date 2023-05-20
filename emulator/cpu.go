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
	StatusNegativeFlag = uint8(0b10000000)
	StatusOverflowFlag = uint8(0b01000000)
	StatusBreak = uint8(0b00010000)
	StatusDecimalMode = uint8(0b00001000)
	StatusInterruptDisable = uint8(0b00000100)
	StatusZeroFlag = uint8(0b00000010)
	StatusCarryFlag = uint8(0b00000001)
)

var OpCodesMap = InitOpCodes()

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
	sp uint8
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
	c.pc = c.bus.ReadU16(0xFFFC)
}

func (c *CPU) fetch() uint8 {
	byte := c.bus.Read(c.pc)
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
	return uint16(hi) << 8 | uint16(lo)
}


func (c *CPU) LoadRom(rom *ROM) {
	c.bus.LoadRom(rom.prgROM)
	c.pc = c.bus.ReadU16(0xFFFC)

	fmt.Printf("Reset Vector: %X\n", c.pc)
}

func (c *CPU) getOperandAddressInMode(addressingMode AddressingMode) uint16 {
	switch addressingMode {
		case Immediate:
			return c.pc
		case ZeroPageX:
			return uint16(c.bus.Read(c.pc) + c.ix)
		case ZeroPageY:
			return uint16(c.bus.Read(c.pc) + c.iy)
		case Absolute:
			return c.bus.ReadU16(c.pc)
		case AbsoluteX:
			return c.bus.ReadU16(c.pc) + uint16(c.ix)
		case AbsoluteY:
			return c.bus.ReadU16(c.pc) + uint16(c.iy)
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
	c.ac = c.bus.Read(c.getOperandAddressInMode(addressingMode))
	c.setZNStatus(c.ac)
}

func (c *CPU) ldx(addressingMode AddressingMode) {
	c.ix = c.bus.Read(c.getOperandAddressInMode(addressingMode))
	c.setZNStatus(c.ix)
}

func (c *CPU) sta(addressingMode AddressingMode) {
	c.bus.Write(c.bus.ReadU16(c.getOperandAddressInMode(addressingMode)), c.ac)
}


func (c *CPU) Step() {

	currentOpcode := c.fetch()

	programCounterStart := c.pc

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
			c.pc = c.bus.ReadU16(c.pc)
		// STA Absolute
		case 0x8D:
			c.sta(Absolute)
		// STA Zero Page X
		case 0x95:
			c.sta(ZeroPageX)
		// STA Absolute X
		case 0x9D:
			c.sta(AbsoluteX)
		// CLD - Clear Decimal Mode
		case 0xD8:
			c.sr &= ^StatusDecimalMode
		// JSR
		case 0x20:
			// address - 1
			c.stackPushU16(c.pc + 2 - 1)
			c.pc = c.bus.ReadU16(c.pc)
		// LDX - Load X Register
		case 0xA2:
			c.ldx(Immediate)
		// TXS - Transfer X to Stack Pointer
		case 0x9A:
			c.sp = c.ix
		// TAX - Transfer Accumulator to X
		case 0xAA:
			c.ix = c.ac
			c.setZNStatus(c.ix)
		// INX - Increment X Register
		case 0xE8:
			c.ix++
			c.setZNStatus(c.ix)
		// BNE
		case 0xD0:
			if (c.sr & StatusZeroFlag) == 0 {
				c.pc += 1 + uint16(c.bus.Read(c.pc))
			}
		// TSX - Transfer Stack Pointer to X
		case 0xBA:
			c.ix = c.sp
			c.setZNStatus(c.ix)
		default:
			panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpcode))
	}

	if c.pc == programCounterStart {
		c.pc += OpCodesMap[currentOpcode].length - 1
	}

}