package main

import (
	"Famicom-emulator/config"
)

// MARK: main関数
func main() {
	rom, config := config.ParseArguments()
	famicom := Famicom{}
	famicom.Init(rom, config)
	famicom.Start()
}
