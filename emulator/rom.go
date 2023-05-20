package emulator

import (
	"bytes"
)

const (
	PRGROMUnitSize = 16384
	CHRROMUnitSize = 8192
)

type ROM struct {
	prgROM []uint8
	chrROM []uint8
}

func CreateROM(rawBytes []uint8) *ROM {

	if !bytes.Equal(rawBytes[:4], []uint8{0x4E, 0x45, 0x53, 0x1A}) {
		panic("Not valid iNES")
	}

	prgROMSize := int(rawBytes[4]) * PRGROMUnitSize
	chrROMSize := int(rawBytes[5]) * CHRROMUnitSize

	skipTrainer := rawBytes[6]&0b1000 != 0

	prgROMStart := 16
	if skipTrainer {
		prgROMStart += 512
	}
	chrROMStart := prgROMStart + prgROMSize

	mapper := (rawBytes[7] & 0b1111_0000) | (rawBytes[6] >> 4)

	if mapper != 0 {
		panic("Only NROM supported rn")
	}

	return &ROM{
		prgROM: rawBytes[prgROMStart : prgROMStart+prgROMSize],
		chrROM: rawBytes[chrROMStart : chrROMStart+chrROMSize],
	}
}
