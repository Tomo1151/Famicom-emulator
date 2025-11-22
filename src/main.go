package main

import (
	"Famicom-emulator/cartridge"
	"Famicom-emulator/config"
)

// MARK: main関数
func main() {
	famicom := Famicom{}
	famicom.Init(
		cartridge.Cartridge{
			// ROM: "../rom/Kirby'sAdventure.nes",
			// ROM: "../rom/SuperMarioBros.nes",
			ROM: "../rom/SuperMarioBros3.nes",
		},
		&config.Config{
			SCALE_FACTOR: 3,
			SOUND_VOLUME: 1.0,
		},
	)
	famicom.Start()
}
