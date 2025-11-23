# Famicom / NES emulator

**Famicom / NES emulator written in Go (1.24.3)**

> [!Tip]
> This emulator was created for my own study purposes, and abstraction has been made at the sacrifice of some accuracy or efficiency.

## Run

To run this program, execute the following command

```shell
cd src && go run .
```

## Dependencies

```
github.com/veandco/go-sdl2
```

To run this program, you need to install SDL2 on your computer. (Windows / mac / linux)

## Directory structure

The directory structure for this program is as follows:

```
├──build: build output dir
├──rom
│   ├──saves: savedata dir
│   └── ***.nes: rom data put here
└──src
     ├──apu
     ├──bus
     ├──cartridge
     │   └──mappers
     ├──cpu
     ├──joypad
     └──ppu
```

Place game ROM data (.nes) under `/rom`.
Save data for games that support saving will be saved in `/rom/saves`.

> [!CAUTION]
> Using games you do not own or illegally obtained ROMs is prohibited.

## Component

This program implements the following components (some of which may not be complete)

- [x] CPU
  - [x] official instructions
  - [x] unofficial instructions
  - [x] nestest.nes: completed
  - [ ] reset
- [x] PPU
  - [x] rough scanline rendering (including scrolling)
  - [ ] accurate sprite 0 hit
  - [ ] accurate scanline rendering emulate
- [x] APU
  - [x] square wave Channel (1 / 2ch)
  - [x] triangle wave Channel (3ch)
  - [x] noise wave Channel (4ch)
  - [x] DMC (5ch)
  - [ ] efficient tick cycle
    - the current implementation has moderately frequent buffer underruns
- [x] Bus
- [x] JoyPad
  - [x] GameController support via SDL2 Gamepad
  - [ ] 2P joypad emulation

## Mappers

The following mappers have been implemented

```
- NROM (mapper 000)
- MMC1: SxROM (mapper 001)
- UxROM (mapper 002)
- CNROM (mapper 003
- MMC3: TxROM (mapper 004)
```

