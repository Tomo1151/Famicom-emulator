package main

import "Famicom-emulator/cartridge"

// MARK: main関数
func main() {
	famicom := Famicom{}
	famicom.Init(cartridge.Cartridge{
		ROM: "../rom/Kirby'sAdventure.nes",
		// ROM: "../rom/SuperMarioBros.nes",
		// ROM: "../rom/SuperMarioBros3.nes",
	})
	famicom.Start()
}
