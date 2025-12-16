# Famicom / NES emulator

**Famicom / NES emulator written in Go (1.24.3)**

> [!Tip]
> This emulator was created for my own study purposes, and abstraction has been made at the sacrifice of some accuracy or efficiency.

## Screenshots

### Game titles

<p align="center">
<img width="352" height="344" alt="スクリーンショット 2025-12-09 11 50 39" src="https://github.com/user-attachments/assets/02619446-3b26-4897-91d4-04b772dd1fe6" />
<img width="352" height="344" alt="スクリーンショット 2025-12-09 11 51 10" src="https://github.com/user-attachments/assets/16896ccf-b52f-49cb-a5b7-827c6d7e4b56" />
<img width="352" height="344" alt="スクリーンショット 2025-12-09 11 51 31" src="https://github.com/user-attachments/assets/a35038de-a99d-417e-958a-3a7ef296c316" />
</p>

### Debug windows

#### Audio visualizer

<p align="center">
     <img width="880" height="459" alt="スクリーンショット 2025-12-09 11 56 28" src="https://github.com/user-attachments/assets/163ba367-8bb3-4e1b-9c8b-c07731d5358f" />
</p>

#### Character ROM viewer

<p align="center">
     <img width="880" height="524" alt="スクリーンショット 2025-12-09 11 57 32" src="https://github.com/user-attachments/assets/f22241d4-ed62-4dda-b73f-038a0da15d8c" />
</p>

#### Nametable viewer

<p align="center">
     <img width="568" height="550" style="text-align: center;" alt="スクリーンショット 2025-12-09 11 57 15" src="https://github.com/user-attachments/assets/fd1229ec-a819-4803-b40c-05ac2f73722a" />
</p>

## Run

To run this program, execute the following command

```shell
cd src && go run .
```

If `"autoLoading"` option in `config.json` is `true`, rom file load from default rom directory automatically.
Otherwise, to launch the game, you'll need to D&D the file into the window after run, or specify the rom file name (relative path from the rom directory) as a startup argument.

> [!Note]
> By default, the last \*.nes file in the rom directory is loaded.

If you use specific Rom file, you can pass Rom path (from default rom directory) first argument.

```shell
cd src && go run . example_rom.nes
```

```shell
cd src && go run . tests/example_rom.nes
```

## Controls

### Gamepad

|               | Famicom / NES (1P/2P) |    1P     |     2P     |
| :------------ | :-------------------: | :-------: | :--------: |
| Button A      |           A           |     K     | / (Slash)  |
| Button B      |           B           |     J     | . (Period) |
| Button Up     |           ↑           |     W     |     G      |
| Button Down   |           ↓           |     S     |     B      |
| Button Right  |           →           |     D     |     N      |
| Button Left   |           ←           |     A     |     V      |
| Button Start  |         Start         |   Enter   |   Enter    |
| Button Select |        Select         | BackSpace | BackSpace  |

> [!Note]
> To change key bindings, edit "key1p/2p" field in `config.json`

### Debug window

|                                                      | Key |
| :--------------------------------------------------- | :-: |
| Exit                                                 | ESC |
| Toggle Fullscreen                                    | F12 |
| Show / Hide nametable viewer                         | F2  |
| Show / Hide CHR ROM viewer                           | F3  |
| Show / Hide audio visualizer                         | F4  |
| Expand debug window                                  |  ↑  |
| Shrink debug window                                  |  ↓  |
| Enable / Disable APU log                             | F10 |
| Enable / Disable CPU log (very processing intensive) | F11 |
| Reset                                                |  Y  |
| Mute / Unmute APU 1ch                                |  1  |
| Mute / Unmute APU 2ch                                |  2  |
| Mute / Unmute APU 3ch                                |  3  |
| Mute / Unmute APU 4ch                                |  4  |
| Mute / Unmute APU 5ch                                |  5  |

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
     ├──ppu
     ├──config: emulator option
     └──ui: emulator / option window
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
  - [x] reset
- [x] PPU
  - [x] rough scanline rendering (including scrolling)
  - [ ] accurate sprite 0 hit
  - [ ] accurate scanline rendering emulate
  - [x] ppu open bus emulate
- [x] APU
  - [x] square wave Channel (1 / 2ch)
  - [x] triangle wave Channel (3ch)
  - [x] noise wave Channel (4ch)
  - [x] DMC (5ch)
  - [ ] expansion audio
- [x] Bus
- [x] JoyPad
  - [x] GameController support via SDL2 Gamepad
  - [x] 2P joypad emulation
  - [x] Joy-Con (R) support
  - [x] Nintendo HVC Controller (1/2) support

## Mappers

The following mappers have been implemented

```
- NROM (mapper 000)
- MMC1: SxROM (mapper 001)
- UxROM (mapper 002)
- CNROM (mapper 003)
- MMC3: TxROM (mapper 004)
```

## Test status

See [TEST_STATUS.md](TEST_STATUS.md) for the latest detailed test results.

The following games have been confirmed to generally play correctly:

- Super Mario Bros. [HVC-SM] (JPN)
- Super Mario Bros.3 [HVC-UM Revision A] (JPN)
- Hoshi no Kirby: Yume no Izumi no Monogatari [HVC-KI] (JPN)

## Trademark Acknowledgment

This is an emulator for the Nintendo Entertainment System (NES) / Family Computer (Famicom).

**"NES," "Nintendo," and "Famicom (Family Computer)" are trademarks of Nintendo Co., Ltd.**

This project is not authorized, approved, or in any way associated with Nintendo. We respect the intellectual property rights of all content owners.
