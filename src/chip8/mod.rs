mod instructions;

use std::io::{self, Write};

pub struct Chip8 {
    pub memory: [u8; 4096],
    pub v: [u8; 16],
    pub i: u16,
    pub pc: u16,
    pub stack: [u16; 16],
    pub sp: u8,
    pub display: [[bool; 64]; 32],
    pub delay_timer: u8,
    pub sound_timer: u8,

    // Keys (keyboard)
    pub keys: [bool; 16],
}

impl Chip8 {
    pub fn new() -> Self {
        let mut chip8 = Chip8 {
            memory: [0; 4096],
            v: [0; 16],
            i: 0,
            pc: 0x200, // Start program counter from 0x200
            stack: [0; 16],
            sp: 0,
            display: [[false; 64]; 32],
            delay_timer: 0,
            sound_timer: 0,
            keys: [false; 16],
        };

        chip8.load_fonts();
        chip8 // Return a chip8 instance
    }

    fn load_fonts(&mut self) {
        let fonts: [u8; 80] = [
            0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
            0x20, 0x60, 0x20, 0x20, 0x70, // 1
            0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
            0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
            0x90, 0x90, 0xF0, 0x10, 0x10, // 4
            0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
            0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
            0xF0, 0x10, 0x20, 0x40, 0x40, // 7
            0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
            0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
            0xF0, 0x90, 0xF0, 0x90, 0x90, // A
            0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
            0xF0, 0x80, 0x80, 0x80, 0xF0, // C
            0xE0, 0x90, 0x90, 0x90, 0xE0, // D
            0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
            0xF0, 0x80, 0xF0, 0x80, 0x80, // F
        ];
        self.memory[0x50..0x50 + fonts.len()].copy_from_slice(&fonts);
    }

    pub fn load_rom(&mut self, rom: &[u8]) {
        self.memory[0x200..0x200 + rom.len()].copy_from_slice(&rom);
    }

    pub fn cycle(&mut self, debug_enabled: bool) -> Result<(), String> {
        // Fetch opcode
        let opcode = (self.memory[self.pc as usize] as u16) << 8
            | (self.memory[self.pc as usize + 1] as u16);

        // Enable terminal debugging if debug_enabled is true
        if debug_enabled {
            let stdout = io::stdout();
            let mut handle = stdout.lock();

            write!(handle, "\x1b[2J\x1b[H").unwrap(); // Clear screen and move cursor

            write!(
                handle,
                "---------------- CHIP-8 DEBUGGER ----------------\n"
            )
            .unwrap();
            write!(
                handle,
                "PC: {:#06X}    I: {:#06X}    Opcode: {:#06X}\n",
                self.pc, self.i, opcode
            )
            .unwrap();
            write!(
                handle,
                "Timers: Delay: {:#04X}    Sound: {:#04X}\n\n",
                self.delay_timer, self.sound_timer
            )
            .unwrap();

            write!(handle, "Registers:\n").unwrap();
            for row in 0..8 {
                for col in 0..2 {
                    let index = row * 2 + col;
                    write!(handle, "V{:X}: {:#04X}  ", index, self.v[index]).unwrap();
                }
                write!(handle, "\n").unwrap();
            }

            write!(handle, "\nStack Frames:\n").unwrap();
            for (i, val) in self.stack.iter().enumerate() {
                write!(handle, "[{}] {:#06X}\n", i, val).unwrap();
            }

            write!(handle, "\nKeys:\n").unwrap();
            for (i, &pressed) in self.keys.iter().enumerate() {
                if pressed {
                    write!(handle, "[{:X}] ", i).unwrap();
                }
            }
            write!(handle, "\n").unwrap();

            write!(
                handle,
                "-------------------------------------------------\n"
            )
            .unwrap();
            handle.flush().unwrap();
        }
        // Increment program counter, now that we have our opcode
        self.pc += 2;

        // Decode and execute
        self.execute(opcode)
    }

    fn execute(&mut self, opcode: u16) -> Result<(), String> {
        // Extract nibbles
        let nnn = opcode & 0x0FFF; // 12-bit address
        let kk = (opcode & 0x00FF) as u8; // 8-bit (byte) value
        let n = (opcode & 0x000F) as u8; // 4-bit nibble
        let x = ((opcode & 0x0F00) >> 8) as usize; // 4-bit value - high byte
        let y = ((opcode & 0x00F0) >> 4) as usize; // 4-bit value - low byte

        // Match opcodes
        match opcode & 0xF000 {
            0x0000 => match kk {
                0xE0 => self.clear_screen(),
                0xEE => self.ret(),
                _ => return Err(format!("Unknown opcode: {:#06X}", opcode)),
            },
            0x1000 => self.jp_addr(nnn),
            0x2000 => self.call_addr(nnn),
            0x3000 => self.se_vx_kk(x, kk),
            0x4000 => self.sne_vx_kk(x, kk),
            0x5000 => self.se_vx_vy(x, y),
            0x6000 => self.ld_vx_kk(x, kk),
            0x7000 => self.add_vx_kk(x, kk),
            0x8000 => match n {
                0x0 => self.ld_vx_vy(x, y),
                0x1 => self.or_vx_vy(x, y),
                0x2 => self.and_vx_vy(x, y),
                0x3 => self.xor_vx_vy(x, y),
                0x4 => self.add_vx_vy(x, y),
                0x5 => self.sub_vx_vy(x, y),
                0x6 => self.shr_vx(x),
                0x7 => self.subn_vx_vy(x, y),
                0xE => self.shl_vx(x),
                _ => return Err(format!("Unknown opcode: {:#06X}", opcode)),
            },
            0x9000 => self.sne_vx_vy(x, y),
            0xA000 => self.ld_i_addr(nnn),
            0xB000 => self.jp_v0_addr(nnn),
            0xC000 => self.rnd_vx_kk(x, kk),
            0xD000 => self.drw_vx_vy_nibble(x, y, n),
            0xE000 => match kk {
                0x9E => self.skp_vx_pressed(x),
                0xA1 => self.skp_vx_npressed(x),
                _ => return Err(format!("Unknown opcode: {:#06}", opcode)),
            },
            0xF000 => match kk {
                0x07 => self.ld_vx_dt(x),
                0x0A => self.ld_vx_k(x),
                0x15 => self.ld_dt_vx(x),
                0x18 => self.ld_st_vx(x),
                0x1E => self.add_i_vx(x),
                0x29 => self.ld_f_vx(x),
                0x33 => self.ld_b_vx(x),
                0x55 => self.ld_i_vx(x),
                0x65 => self.ld_vx_i(x),
                _ => return Err(format!("Unknown opcode: {:#06}", opcode)),
            },
            _ => return Err(format!("Unknown opcode: {:#06}", opcode)),
        }

        Ok(())
    }
}
