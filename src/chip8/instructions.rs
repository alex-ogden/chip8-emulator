use crate::chip8::Chip8;

impl Chip8 {
    // 00E0 - CLS: Clear the display
    pub fn clear_screen(&mut self) {
        self.display = [[false; 64]; 32];
    }

    // 00EE - RET: Return from a subroutine
    pub fn ret(&mut self) {
        self.sp -= 1;
        self.pc = self.stack[self.sp as usize];
    }

    // 1nnn - JP addr: Jump to location nnn
    pub fn jp_addr(&mut self, addr: u16) {
        self.pc = addr;
    }

    // 2nnn - CALL addr: Call subroutine at nnn
    pub fn call_addr(&mut self, addr: u16) {
        self.stack[self.sp as usize] = self.pc;
        self.sp += 1;
        self.pc = addr;
    }

    // 3xkk - SE Vx, byte: Skip next instruction if Vx = kk
    pub fn se_vx_kk(&mut self, x: usize, kk: u8) {
        if self.v[x] == kk {
            self.pc += 2;
        }
    }

    // 4xkk - SNE Vx, byte: Skip next instruction if Vx != kk
    pub fn sne_vx_kk(&mut self, x: usize, kk: u8) {
        if self.v[x] != kk {
            self.pc += 2;
        }
    }

    // 5xy0 - SE Vx, Vy: Skip next instruction if Vx = Vy
    pub fn se_vx_vy(&mut self, x: usize, y: usize) {
        if self.v[x] == self.v[y] {
            self.pc += 2;
        }
    }

    // 6xkk - LD Vx, byte: Set Vx = kk
    pub fn ld_vx_kk(&mut self, x: usize, kk: u8) {
        self.v[x] = kk;
    }

    // 7xkk - ADD Vx, byte: Set Vx = Vx + kk
    pub fn add_vx_kk(&mut self, x: usize, kk: u8) {
        self.v[x] = self.v[x].wrapping_add(kk);
    }

    // 8xy0 - LD Vx, Vy: Set Vx = Vy
    pub fn ld_vx_vy(&mut self, x: usize, y: usize) {
        self.v[x] = self.v[y];
    }

    // 8xy1 - OR Vx, Vy: Set Vx = Vx OR Vy
    pub fn or_vx_vy(&mut self, x: usize, y: usize) {
        self.v[x] |= self.v[y];
        self.v[0xF] = 0;
    }

    // 8xy2 - AND Vx, Vy: Set Vx = Vx AND Vy
    pub fn and_vx_vy(&mut self, x: usize, y: usize) {
        self.v[x] &= self.v[y];
        self.v[0xF] = 0;
    }

    // 8xy3 - XOR Vx, Vy: Set Vx = Vx XOR Vy
    pub fn xor_vx_vy(&mut self, x: usize, y: usize) {
        self.v[x] ^= self.v[y];
        self.v[0xF] = 0;
    }

    // 8xy4 - ADD Vx, Vy: Set Vx = Vx + Vy, set VF = carry
    pub fn add_vx_vy(&mut self, x: usize, y: usize) {
        let (result, overflow) = self.v[x].overflowing_add(self.v[y]);
        self.v[x] = result;
        self.v[0xF] = if overflow { 1 } else { 0 };
    }

    // 8xy5 - SUB Vx, Vy: Set Vx = Vx - Vy, set VF = NOT borrow
    pub fn sub_vx_vy(&mut self, x: usize, y: usize) {
        let (result, borrow) = self.v[x].overflowing_sub(self.v[y]);
        self.v[x] = result;
        self.v[0xF] = if borrow { 0 } else { 1 };
    }

    // 8xy6 - SHR Vx: Set Vx = Vx SHR 1
    pub fn shr_vx(&mut self, x: usize) {
        self.v[0xF] = self.v[x] & 0x1;
        self.v[x] >>= 1;
    }

    // 8xy7 - SUBN Vx, Vy: Set Vx = Vy - Vx, set VF = NOT borrow
    pub fn subn_vx_vy(&mut self, x: usize, y: usize) {
        let (result, borrow) = self.v[y].overflowing_sub(self.v[x]);
        self.v[x] = result;
        self.v[0xF] = if borrow { 0 } else { 1 };
    }

    // 8xyE - SHL Vx: Set Vx = Vx SHL 1
    pub fn shl_vx(&mut self, x: usize) {
        self.v[0xF] = (self.v[x] & 0x80) >> 7;
        self.v[x] <<= 1;
    }

    // 9xy0 - SNE Vx, Vy: Skip next instruction if Vx != Vy
    pub fn sne_vx_vy(&mut self, x: usize, y: usize) {
        if self.v[x] != self.v[y] {
            self.pc += 2;
        }
    }

    // Annn - LD I, addr: Set I = nnn
    pub fn ld_i_addr(&mut self, addr: u16) {
        self.i = addr;
    }

    // Bnnn - JP V0, addr: Jump to location nnn + V0
    pub fn jp_v0_addr(&mut self, addr: u16) {
        self.pc = addr + self.v[0] as u16;
    }

    // Cxkk - RND Vx, byte: Set Vx = random byte AND kk
    pub fn rnd_vx_kk(&mut self, x: usize, kk: u8) {
        let random_byte: u8 = rand::random();
        self.v[x] = random_byte & kk;
    }

    // Dxyn - DRW Vx, Vy, // Dxyn - DRW Vx, Vy, nibble: Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision
    pub fn drw_vx_vy_nibble(&mut self, x: usize, y: usize, n: u8) {
        self.v[0xF] = 0;

        for byte in 0..n {
            let y_coord = ((self.v[y] as usize + byte as usize) % 32) as usize;
            let sprite_byte = self.memory[(self.i + byte as u16) as usize];

            for bit in 0..8 {
                let x_coord = ((self.v[x] as usize + bit as usize) % 64) as usize;
                let pixel = (sprite_byte >> (7 - bit)) & 1;

                if pixel == 1 {
                    if self.display[y_coord][x_coord] {
                        self.v[0xF] = 1; // Collision detected
                    }
                    self.display[y_coord][x_coord] ^= true;
                }
            }
        }
    }

    // Ex9E - SKP Vx: Skip next instruction if key with the value of Vx is pressed
    pub fn skp_vx_pressed(&mut self, x: usize) {
        let key = self.v[x] as usize;
        if key < 16 && self.keys[key] {
            self.pc += 2;
        }
    }

    // ExA1 - SKNP Vx: Skip next instruction if key with the value of Vx is not pressed
    pub fn skp_vx_npressed(&mut self, x: usize) {
        let key = self.v[x] as usize;
        if key >= 16 || !self.keys[key] {
            self.pc += 2;
        }
    }

    // Fx07 - LD Vx, DT: Set Vx = delay timer value
    pub fn ld_vx_dt(&mut self, x: usize) {
        self.v[x] = self.delay_timer;
    }

    // Fx0A - LD Vx, K: Wait for a key press, store the value of the key in Vx
    pub fn ld_vx_k(&mut self, x: usize) {
        let mut key_pressed = false;

        for (i, &key) in self.keys.iter().enumerate() {
            if key {
                self.v[x] = i as u8;
                key_pressed = true;
                break;
            }
        }

        // If no key pressed, decrement PC to repeat this instruction
        if !key_pressed {
            self.pc -= 2;
        }
    }

    // Fx15 - LD DT, Vx: Set delay timer = Vx
    pub fn ld_dt_vx(&mut self, x: usize) {
        self.delay_timer = self.v[x];
    }

    // Fx18 - LD ST, Vx: Set sound timer = Vx
    pub fn ld_st_vx(&mut self, x: usize) {
        self.sound_timer = self.v[x];
    }

    // Fx1E - ADD I, Vx: Set I = I + Vx
    pub fn add_i_vx(&mut self, x: usize) {
        self.i = self.i.wrapping_add(self.v[x] as u16);
    }

    // Fx29 - LD F, Vx: Set I = location of sprite for digit Vx
    pub fn ld_f_vx(&mut self, x: usize) {
        let digit = self.v[x] & 0x0F; // Only use lower 4 bits (0-F)
        self.i = 0x50 + (digit as u16 * 5); // Each font sprite is 5 bytes
    }

    // Fx33 - LD B, Vx: Store BCD representation of Vx in memory locations I, I+1, and I+2
    pub fn ld_b_vx(&mut self, x: usize) {
        let value = self.v[x];
        self.memory[self.i as usize] = value / 100; // Hundreds
        self.memory[self.i as usize + 1] = (value / 10) % 10; // Tens
        self.memory[self.i as usize + 2] = value % 10; // Ones
    }

    // Fx55 - LD [I], Vx: Store registers V0 through Vx in memory starting at location I
    pub fn ld_i_vx(&mut self, x: usize) {
        for i in 0..=x {
            self.memory[self.i as usize + i] = self.v[i];
        }
    }

    // Fx65 - LD Vx, [I]: Read registers V0 through Vx from memory starting at location I
    pub fn ld_vx_i(&mut self, x: usize) {
        for i in 0..=x {
            self.v[i] = self.memory[self.i as usize + i];
        }
    }
}
