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
	// Only for JMP
	Indirect
	IndirectX
	IndirectY
)

const (
	StatusNegativeFlag     = uint8(0b10000000)
	StatusOverflowFlag     = uint8(0b01000000)
	StatusBreak            = uint8(0b00010000)
	StatusDecimalMode      = uint8(0b00001000)
	StatusInterruptDisable = uint8(0b00000100)
	StatusZeroFlag         = uint8(0b00000010)
	StatusCarryFlag        = uint8(0b00000001)
)

var OpCodesMap = InitOpCodes()

type CPU struct {
	bus   Bus
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
	lo := uint8((value & 0xff) >> 8)
	c.stackPush(hi)
	c.stackPush(lo)
}

func (c *CPU) stackPopU16() uint16 {
	lo := c.stackPop()
	hi := c.stackPop()

	return uint16(hi)<<8 | uint16(lo)
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
	case ZeroPage:
		return uint16(c.bus.Read(c.pc))
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
	case Indirect:
		// Indirect (JMP)
		// The next 16-bit address is used to get the actual 16-bit address. This instruction has
		// a bug in the original hardware. If the lo byte is 0xFF, the hi byte would cross a page
		// boundary. However, this doesn't work correctly on the original hardware and instead
		// wraps back around to 0.
		address := c.bus.ReadU16(c.pc)
		if (address & 0xFF) == 0xFF {
			lo := c.bus.Read(address)
			hi := c.bus.Read(address & 0xFF00)
			return uint16(hi)<<8 | uint16(lo)
		}
		return c.bus.ReadU16(address)
	case IndirectX:
		return c.bus.ReadU16(c.bus.ReadU16(c.pc) + uint16(c.ix))
	case IndirectY:
		return c.bus.ReadU16(c.bus.ReadU16(c.pc)) + uint16(c.iy)
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

func (c *CPU) ldy(addressingMode AddressingMode) {
	c.iy = c.bus.Read(c.getOperandAddressInMode(addressingMode))
	c.setZNStatus(c.iy)
}

func (c *CPU) cmp(addressingMode AddressingMode) {
	value := c.bus.Read(c.getOperandAddressInMode(addressingMode))

	if c.ac >= value {
		c.sr |= StatusCarryFlag
	} else if c.ac == value {
		c.sr |= StatusZeroFlag
	}

	if ((c.ac - value) >> 7) == 0x1 {
		c.sr |= StatusNegativeFlag
	}
}

func (c *CPU) bit(addressingMode AddressingMode) {
	value := c.bus.Read(c.getOperandAddressInMode(addressingMode))
	if (value & c.ac) == 0 {
		c.sr |= StatusZeroFlag
	} else {
		c.sr &= ^StatusNegativeFlag
	}
}

func (c *CPU) inc(addressingMode AddressingMode) {
	address := c.getOperandAddressInMode(addressingMode)
	value := c.bus.Read(address) + 1
	c.bus.Write(address, value)
	c.setZNStatus(value)
}

func (c *CPU) ora(addressingMode AddressingMode) {
	c.ac |= c.bus.Read(c.getOperandAddressInMode(addressingMode))
}

func (c *CPU) sta(addressingMode AddressingMode) {
	c.bus.Write(c.bus.ReadU16(c.getOperandAddressInMode(addressingMode)), c.ac)
}

func (c *CPU) stx(addressingMode AddressingMode) {
	c.bus.Write(c.bus.ReadU16(c.getOperandAddressInMode(addressingMode)), c.ix)
}

func (c *CPU) Step() {

	currentOpcode := c.fetch()

	programCounterStart := c.pc

	fmt.Printf("Current OpCode: %X\n", currentOpcode)
	fmt.Printf("Current PC: %X\n", programCounterStart-1)

	switch currentOpcode {
	// LDA Immediate
	case 0xA9:
		c.lda(Immediate)
	// LDA Indirect Y
	case 0xB1:
		c.lda(IndirectY)
	// LDA Zero Page
	case 0xA5:
		c.lda(ZeroPage)
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
	// DEX - Decrement X Register
	case 0xCA:
		c.ix--
		c.setZNStatus(c.ix)
	// RTS - Return from Subroutine
	case 0x60:
		c.pc = c.stackPopU16() + 1
	// LDY - Load Y Register Immediate
	case 0xA0:
		c.ldy(Immediate)
	// BIT - BIT Test Absolute
	case 0x2C:
		c.bit(Absolute)
	// BMI - Branch if Minus
	case 0x30:
		if (c.sr & StatusNegativeFlag) != 0 {
			c.pc += 1 + uint16(c.bus.Read(c.pc))
		}
	// TYA - Transfer Y to Accumulator
	case 0x98:
		c.ac = c.iy
		c.setZNStatus(c.ac)
	// ORA Absolute
	case 0x0D:
		c.ora(Absolute)
	// STX Zero Page
	case 0x86:
		c.stx(ZeroPage)
	// STA Zero Page
	case 0x85:
		c.sta(ZeroPage)
	// BPL
	case 0x10:
		if (c.sr & StatusNegativeFlag) == 0 {
			c.pc += 1 + uint16(c.bus.Read(c.pc))
		}
	// LDX Zero Page
	case 0xA6:
		c.ldx(ZeroPage)
	// PHA
	case 0x48:
		c.stackPush(c.ac)
	// PLA
	case 0x68:
		c.ac = c.stackPop()
		c.setZNStatus(c.ac)
	// STX Absolute
	case 0x8E:
		c.stx(Absolute)
	// TAY
	case 0xA8:
		c.iy = c.ac
		c.setZNStatus(c.iy)
	// PLP
	case 0x28:
		c.sr = c.stackPop()
	// BEQ
	case 0xF0:
		if (c.sr & StatusZeroFlag) != 0 {
			c.pc += 1 + uint16(c.bus.Read(c.pc))
		}
	// INC Zero Page
	case 0xE6:
		c.inc(ZeroPage)
	// JMP Indirect
	case 0x6C:
		c.pc = c.getOperandAddressInMode(Indirect)
	// BRK
	case 0x00:
		c.stackPushU16(c.pc)
		c.stackPush(c.sr)
		c.pc = c.bus.ReadU16(0xFFFE)
		c.sr |= StatusBreak
	// RTI
	case 0x40:
		c.sr = c.stackPop()
		c.pc = c.stackPopU16()
	// CMP Immediate
	case 0xC9:
		c.cmp(Immediate)
	default:
		panic(fmt.Sprintf("Unknown Instruction: 0x%X at 0x%X\n", currentOpcode, c.pc-1))
	}

	if c.pc == programCounterStart {
		c.pc += OpCodesMap[currentOpcode].length - 1
	}

}
