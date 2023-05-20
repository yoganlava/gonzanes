package main

import (
	"go-nes/emulator"
	"os"
)

func main() {
	cpu := emulator.Init()
	file, err := os.ReadFile("cpu_dummy_writes_oam.nes")
	if err != nil {
		panic("no File")
	}
	rom := emulator.CreateROM(file)

	cpu.LoadRom(rom)

	for {
		cpu.Step()
	}
}
