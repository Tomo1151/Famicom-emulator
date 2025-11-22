package main

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/config"
)

// MARK: main関数
func main() {
	famicom := Famicom{}
	cfg := config.New()
	famicom.Init(cartridge.Cartridge{
		// ROM: "../rom/Kirby'sAdventure.nes",
		// ROM: "../rom/SuperMarioBros.nes",
		ROM: "../rom/SuperMarioBros3.nes",
	}, cfg)
	famicom.Start()
}
