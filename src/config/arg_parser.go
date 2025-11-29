package config

import (
	"Famicom-emulator/cartridge"
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// MARK: コマンドライン引数をパースしてROMとコンフィグを生成
func ParseArguments() (cartridge.Cartridge, *Config) {
	var (
		rom = flag.String("rom", defaultRomName(), "Rom file name in /rom.")
	)

	flag.Parse()

	fmt.Println("Load ROM file:", *rom)

	return cartridge.Cartridge{
		ROM: filepath.Join("..", "rom", *rom),
	}, DefaultConfig
}

// MARK: デフォルトのROMファイル名を取得
func defaultRomName() string {
	// rom/*.nes にマッチする一番最後のROMを返す
	pattern := filepath.Join("..", "rom", "*.nes")
	matches, err := filepath.Glob(pattern)

	if err != nil || len(matches) == 0 {
		return ""
	}

	sort.Slice(matches, func(i, j int) bool {
		a := strings.ToLower(filepath.Base(matches[i]))
		b := strings.ToLower(filepath.Base(matches[j]))
		return a < b
	})

	return filepath.Base(matches[len(matches)-1])
}
