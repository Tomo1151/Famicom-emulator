# Test status

- **S : success**
- **F : fail**
- **P : partialy success**
- **U : unverifiable**
- **NI: not implemented**

## APU

- apu_mixer
  - dmc: F
    > don't become silent
  - noise: F
    > another sound is playing
  - square: S
  - triangle: P
    > don't become silent
- apu_reset
  - 4015_cleared: S
  - 4017_timing: F
    > Delay after effective \$4017  
    > write: 0  
    > Frame IRQ flag should be set sooner after power/reset  
    > Failed #3
  - 4017_written: F
    > At power, \$4017 should be written with \$00  
    > Failed #2
  - irq_flag_cleared: S
  - len_ctrs_enabled: S
  - works_immediately: F
    > At power, writes should work immediately  
    > Failed #2
- apu_test
  - rom_singles
    - 1-len_ctr: S
    - 2-len_table: S
    - 3-irq_flag: F
      > Writing \$00 or \$80 to \$4017  
      > shouldn't affect flag  
      > Failed #6
    - 4-jitter: F
      > Frame irq is set too soon  
      > Failed #2
    - 5-len_timing: F
      > Channel: 0  
      > First length of mode 0 is too soon  
      > Failed #2
    - 6-irq_flag_timing: F
      > Flag first set too soon  
      > Failed #2
    - 7-dmc_basics: F
      > Channel: 0  
      > DMC isn't working well enough to test further  
      > Failed #2
    - 8-dmc_rates: U
- blargg_apu_2005.07.30
  - 01.len_ctr: F
    > 2: Problem with length counter load or \$4015
  - 02.len_table: F
    > 2: Prints four bytes \$II \$ee \$cc \$02 that indicate the length load value written (ll), the value that the emulator uses (\$ee), and the correct value (\$cc).
  - 03.irq_flag: F
    > 6: Writing \$00 or \$80 to \$4017 doesn't affect flag
  - 04.clock_jitter: F
    > 2: Frame irq is set too soon
  - 05.len_timing_mode0: F
    > 2: First length is clocked too soon
  - 06.len_timing_mode1: F
    > 2: First length is clocked too soon
  - 07.irq_flag_timing: F
    > 2: Flag first set too soon
  - 08.irq_timing: F
    > 2: Too soon
  - 09.reset_timing: F
    > 4: Fourth step occurs too late
  - 10.len_halt_timing: F
    > 2: Length shouldn't be clocked when halted at 14914
  - 11.len_reload_timing: F
    > 2: Reload just before length clock should work normally
- dmc_dma_during_read4
  - dma_2007_read: U
    > unable to verify because the screen is not displayed
  - dma_2007_write: U
    > unable to verify because the screen is not displayed
  - dma_4016_read: U
    > unable to verify because the screen is not displayed
  - double_2007_read: U
    > unable to verify because the screen is not displayed
  - read_write_2007: F
- dmc_tests
  - buffer_retained: S?
  - latency: S?
  - status: U
    > unable to verify because the screen is not displayed
  - status_irq: S?
- dpcmletterbox
  - dpcmletterbox: F
- volume_tests
  - volumes: S?

## CPU

- blargg_nes_cpu_test5
  - cpu: F
    > unsupported mirroring mode
  - official: F
    > unsupported mirroring mode
- branch_timing_tests
  - 1.Branch_Basics: P
  - 2.Backward_Branch: F
    > 5: Branch from \$E5FE to \$E5FD is too short
  - 3.Forward_Branch: F
    > 5: Branch from \$E6FD to \$E700 is too short
- cpu_dummy_reads
  - cpu_dummy_reads: F
    > 3
- cpu_dummy_writes
  - cpu_dummy_writes_oam: U
    > \$2002 must be mirrored every 8 bytes to \$3FFA  
    > Error 2
  - cpu_dummy_writes_ppumem: F
    > panic: attempt to write to PPU Status register (specification investigation required)
- cpu_exec_space
  - test_cpu_exec_space_apu: F
    > panic: attempt to read from OAM Data register (specification investigation required)
  - test_cpu_exec_space_ppuio: F
    > An RTS opcode should still do a dummy fetch of the next opcode. (The same goes for all one-byte opcodes, really).  
    > Failed #5
- cpu_interrupts_v2

  - rom_singles

    - 1-cli_latency: F
      > Exactly one instruction after CLI should execute before IRQ is taken  
      > Failed #4
    - 2-nmi_and_brk: F

      > NMI BRK 00  
      > 07 00 01  
      > 07 00 01  
      > 07 00 01  
      > 07 00 01  
      > 07 00 01  
      > 07 00 01  
      > 06 00 01  
      > 06 00 01  
      > 06 00 01  
      > 06 00 01  
      > 06 00 01
      >
      > 9AD4D079  
      > Failed

    - 3-nmi_and_irq: F

      > NMI BRK  
      > 02 00  
      > 07 02  
      > 07 03  
      > 07 03  
      > 07 03  
      > 07 02  
      > 07 03  
      > 07 02  
      > 05 02  
      > 05 03  
      > 05 03  
      > 05 03
      >
      > DAF06684  
      > Failed

    - 4-irq_and_dma: F

      > 53 +0  
      > 53 +1  
      > 53 +2  
      > 53 +3  
      > 0 +4  
      > 0 +5  
      > 0 +6  
      > 0 +7  
      > 0 +8  
      > 1 +9  
      > 1 +10  
      > 2 +11  
      > 2 +12  
      > 4 +13  
      > ...  
      > 53 +524  
      > 53 +525  
      > 53 +526  
      > 53 +527
      >
      > 54D414C8  
      > Failed

    - 5-branch_delays_irq: F
      > test_jmp  
      > T+ CK PC  
      > 00 02 0E  
      > 01 02 0E  
      > 02 04 03  
      > 03 03 03  
      > 04 03 03  
      > 05 03 03  
      > 06 02 03  
      > 07 02 03  
      > 08 02 04  
      > 09 02 04
      >
      > 71833847  
      > Failed

- cpu_reset
  - ram_after_reset: S
  - registers: S
- cpu_timing_test6
  - cpu_timing_test: S
- instr_misc
  - rom_singles
    - 01-abs_x_wrap: S
    - 02-branch_wrap: S
    - 03-dummy_reads: F
      > Test requires \$2002 mirroring every 8 bytes to \$3FFA  
      > Failed #2
    - 04-dummy_reads_apu: F
      > 1D 19 11 3D 39 31 5D 59 51 7D  
      > 79 71 9D 99 91 BD B9 B1 DD D9  
      > D1 FD F9 F1 1E 3E 5E 7E DE FE  
      > BC BE  
      > Official opcodes failed  
      > Failed #2
- instr_test-v3
  - all_instrs: F
    > unsupported mirroring mode
  - official_only: F
    > unsupported mirroring mode
  - rom_singles
    - 01-implied: S
    - 02-immediate: F
      > 6B ARR #n  
      > AB ATX #n  
      > Failed
    - 03-zero_page: S
    - 04-zp_xy: S
    - 05-absolute: S
    - 06-abs_xy: F
      > 9C SYA abs, X  
      > 9E SXA abs, Y  
      > Failed
    - 07-ind_x: S
    - 08-ind_y: S
    - 09-branches: S
    - 10-stack: S
    - 11-jmp_jsr: S
    - 12-rts: S
    - 13-rti: S
    - 14-brk: F
      > panic: attempt to read from PPU Address register
    - 15-special: F
      > panic: unsupported: read program ROM on NROM
- instr_timing
  - instr_timing: F
    > unsupported mirroring mode
- other
  - nestest
    - Normal ops
      - Branch tests: S
      - Flag tests: S
      - Immediate tests: S
      - Implied tests: S
      - Stack tests: S
      - Accumulator tests: S
      - (Indirect, X), tests: S
      - Zeropage tests: S
      - Absolute tests: S
      - (Indirect), Y tests: S
      - Absolute, Y tests: S
      - Zeropage, X tests: S
      - Absolute, X tests: S
    - Invalid ops
      - NOP tests: S
      - LAX tests: S
      - SAX tests: S
      - SBC tets: S
      - DCP tests: S
      - ISB tests: S
      - SLO tets: S
      - RLA tets: S
      - SRE tests: S
      - RRA tests: S
- nes_instr_test
  - rom_singles
    - 01-implied: F
      > unsupported read program RAM on NROM
    - 02-immediate: F
      > unsupported read program RAM on NROM
    - 03-zero_page: F
      > unsupported read program RAM on NROM
    - 04-zp_xy: F
      > unsupported read program RAM on NROM
    - 05-absolute: F
      > unsupported read program RAM on NROM
    - 06-abs_xy: F
      > unsupported read program RAM on NROM
    - 07-ind_x: F
      > unsupported read program RAM on NROM
    - 08-ind_y: F
      > unsupported read program RAM on NROM
    - 09-branches: F
      > unsupported read program RAM on NROM
    - 10-stack: F
      > unsupported read program RAM on NROM
    - 11-special: F
      > unsupported read program RAM on NROM

## Mapper

- mmc3_irq_tests
  - 1.Clocking: F
    > Should decrement when A12 is toggled via \$2006  
    > Failed #3
  - 2.Details: F
    > Counter isn't working when reloaded with 255  
    > Failed #2
  - 3.A12_clocking: F
    > Should be clocked when A12 changes to 1 via \$2006 write  
    > Failed #4
  - 4.Scanline_timing: F
    > Scanline 0 time is too late  
    > Failed #3
  - 5.MMC3_rev_A: F
    > IRQ should be set when reloading to 0 after clear  
    > Failed #2
  - 6.MMC3_rev_B: F
    > Should reload and set IRQ every clock when reload is 0  
    > Failed #2
- mmc3_test
  - 1-clocking: F
    > Should decrement when A12 is toggled via PPUADDR  
    > Failed #3
  - 2-details: F
    > Counter isn't working when reloaded with 255  
    > Failed #2
  - 3-A12_clocking: F
    > Should be clocked when A12 changes to 1 via PPUADDR write  
    > Failed #4
  - 4-scanline_timing: F
    > Scanline 0 IRQ should occur sooner when \$2000=\$08  
    > Failed #3
  - 5-MMC3: F
    > Should reload and set IRQ every clock when reload is 0  
    > Failed #2
  - 6-MMC6: F
    > IRQ should be set when reloading to 0 after clear  
    > Failed #2

## PPU

- blargg_ppu_tests_2005.09.15b
  - palette_ram: S
  - power_up_palette: F
    > 2: Palette differs from table
  - sprite_ram: S
  - vbl_clear_time: F
    > 3: VBL flag cleared too late
  - vram_access: F
    > 6: Palette read should also read VRAM into read buffer
- nmi_sync
  - demo_ntsc: F
  - demo_pal: F
- oam_read
  - oam_read: S
- oam_stress
  - oam_stress: F
    > ```
    > ------*---*---*-
    > --*---*---*---*-
    > --*---*---*---*-
    > --*---*---*---*-
    > --*---*---*---*-
    > --*---*---*---*-
    > --*---*---*---*-
    > ```
    >
    > 59916E5B  
    > Failed
- ppu_open_bus
  - ppu_open_bus: S
- ppu_vbl_nmi

  - rom_singles

    - 01-vbl_basics.nes: F
      > panic: attemp to write to PPU Status register
    - 02-vbl_set_time: F

      > T+ 1 2  
      > 00 - V  
      > 01 - V  
      > 02 - V  
      > 03 - V  
      > 04 - V  
      > 05 V -  
      > 06 V -  
      > 07 V -  
      > 08 V -
      >
      > 4103C340  
      > Failed

    - 03-vbl_clear_time: F

      > 00 V  
      > 01 V  
      > 02 V  
      > 03 V  
      > 04 V  
      > 05 V  
      > 06 V  
      > 07 V  
      > 08 V
      >
      > F9DC460  
      > Failed

    - 04-nmi_control: F
      > Immediate occurence should be after NEXT instruction  
      > Failed #11
    - 05-nmi_timing: F

      > 00 2  
      > 01 1  
      > 02 1  
      > 03 1  
      > 04 1  
      > 05 1  
      > 06 1  
      > 07 0  
      > 08 0  
      > 09 0
      >
      > D688789E  
      > Failed

    - 06-suppression: F

      > 00 - N  
      > 01 - N  
      > 02 - N  
      > 03 - N  
      > 04 - N  
      > 05 V N  
      > 06 V N  
      > 07 V N  
      > 08 V N  
      > 09 V N
      >
      > 3FE15516  
      > Failed

    - 07-nmi_on_timing: F

      > 00 N  
      > 01 N  
      > 02 N  
      > 03 N  
      > 04 N  
      > 05 N  
      > 06 N  
      > 07 N  
      > 08 N
      >
      > 6BA71A6F  
      > Failed

    - 08-nmi_off_timing: F

      > 03 -  
      > 04 -  
      > 05 N  
      > 06 N  
      > 07 N  
      > 08 N  
      > 09 N  
      > 0A N  
      > 0B N  
      > 0C N
      >
      > 4CC88927  
      > Failed

    - 09-even_odd_frames: F
      > 00 00  
      > Pattern ---BB should skip 1 clock  
      > Failed #3
    - 10-even_odd_timing: F
      > 09  
      > Clock is skipped too soon, relative to enabling BG  
      > Failed #2

- scanline
  - scanline: P
    > the following section is displayed incorrectly
    > This third area uses \$2005/\$2006 to update the VRAM addres at the proper locations.
- scrolltest
  - scroll: F
    > panic: unsupported mirroring mode
- sprdma_and_dmc_dma
  - sprdma_and_dmc_dma_512: U
    > unable to verify because the screen is not displayed
  - sprdma_and_dmc_dma: U
    > unable to verify because the screen is not displayed
    > T+ clock (decimal) ?
- sprite_hit_tests_2005.10.05
  - 01.basics: F
    > 4: Should miss when background rendering is off
  - 02.alignment: F
    > 3: Sprite should miss left side of bg tile
  - 03.corners: S
  - 04.flip: S
  - 05.left_clip: F
    > 2: Should miss when entirely n left-edge clipping
  - 06.right_edge: F
    > 2: Should always miss when X = 255
  - 07.screen_bottom: F
    > 3: Can hit when Y < 239
  - 08.double_height: F
    > 2: Lower sprite tile should miss bottom of bg tile
  - 09.timing_basics: F
    > 3: Upper-left corner too late
  - 10.timing_order: F
    > 3: Upper-left corner too late
  - 11.edge_timing: S
- sprite_overflow_tests
  - 1.Basics: F
    > 5: Should be cleared at the end of VBL
  - 2.Details: F
    > 3: Disabling rendering shouldn't clear flag
  - 3.Timing: F
    > 2: Cleared too late
  - 4.Obscure: F
    > 2: Checks that second byte of sprite #10 is treated as its Y
  - 5.Emulator: F
    > 3: Disabling rendering didn't recalculate flag time
- tvpassfail
  - tv: P
    > incorrect aspect ratio
- vbl_nmi_timing
  - 1.frame_basics: F
    > 5: PPU frame with BG enabled is too long
  - 2.vbl_timing: F
    > 8: Reading 1 PPU clock before VBL should suppress setting
  - 3.even_odd_frames: F
    > 3: Pattern BB--- should skip 1 clock
  - 4.vbl_clear_timing: F
    > 5: Cleared 3 or more PPU clock too late
  - 5.nmi_suppression: F
    > 3: Reading flag when it's set should suppress NMI
  - 6.nmi_disable: F
    > 2: NMI shouldn't occur when disabled 0 PPU clock after VBL
  - 7.nmi_timing: F
    > 2: NMI occurred 3 or more PPU clocks too early

## Miscellaneous

- PaddleTest3
  - PaddleTest: F
- read_joy3
  - test_buttons: S
  - thorough_test: U
    > unable to verify because the screen is not displayed

## Demos that require accuracy

- full_palette
  - flowing_palette: F
    > panic: index out of range [241] with length 64 at PPU.ClearLineBuffer(...) from ppu.RenderScanlineToCanvas
  - full_palette_smooth: F
    > only blue colors are displayed
  - full_palette: F
    > only blue colors are displayed
