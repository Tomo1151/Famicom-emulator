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
	config := LoadFromFile()

	flag.Parse()
	rom := flag.Arg(0)
	if len(rom) == 0 {
		rom = defaultRomPath()
	}

	fmt.Println("Load ROM file:", rom)

	return cartridge.Cartridge{
		ROM: filepath.Join("..", "rom", rom),
	}, config
}

// MARK: デフォルトのROMファイル名を取得
func defaultRomPath() string {
	// rom/*.nes にマッチするROMのパスを取得
	pattern := filepath.Join("..", "rom", "*.nes")
	matches, err := filepath.Glob(pattern)

	if err != nil || len(matches) == 0 {
		return ""
	}

	// 大文字小文字を揃えて辞書順に並べる
	sort.Slice(matches, func(i, j int) bool {
		a := strings.ToLower(filepath.Base(matches[i]))
		b := strings.ToLower(filepath.Base(matches[j]))
		return a < b
	})

	// 取得されたROMの中で一番最後のものを返す
	return filepath.Base(matches[len(matches)-1])
}
