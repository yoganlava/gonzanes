package emulator

type OpCode struct {
	code uint8
	length uint16
}

func InitOpCodes() map[uint8]OpCode {
	return map[uint8]OpCode{
		0xA9: {
			0xA9,
			2,
		},
		0x78: {
			0x78,
			1,
		},
		0x4C: {
			0x4C,
			1,
		},
		0x8D: {
			0x8D,
			3,
		},
		0x95: {
			0x95,
			2,
		},
		0xD8: {
			0xD8,
			1,
		},
		0x20: {
			0x20,
			3,
		},
		0xA2: {
			0xA2,
			2,
		},
		0x9A: {
			0x9A,
			1,
		},
		0xAA: {
			0xAA,
			1,
		},
		0x9D: {
			0x9D,
			3,
		},
		0xE8: {
			0xE8,
			1,
		},
		0xD0: {
			0xD0,
			2,
		},
		0xBA: {
			0xBA,
			1,
		},
	}
}