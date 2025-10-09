package chip8

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

var fontSet = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, //0
	0x20, 0x60, 0x20, 0x20, 0x70, //1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
	0x90, 0x90, 0xF0, 0x10, 0x10, //4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
	0xF0, 0x10, 0x20, 0x40, 0x40, //7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
	0xF0, 0x90, 0xF0, 0x90, 0x90, //A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
	0xF0, 0x80, 0x80, 0x80, 0xF0, //C
	0xE0, 0x90, 0x90, 0x90, 0xE0, //D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
	0xF0, 0x80, 0xF0, 0x80, 0x80, //F
}

// CHIP-8 implementation struct
type Chip8 struct {
	display [32][64]uint8 // 64x32 res display

	memory [4096]uint8 // 4K of memory
	vx     [16]uint8   // V registers (V0->VF)
	key    [16]uint8   // Input keys
	stack  [16]uint16  // The program stack

	pc uint16 // Holds the program counter
	oc uint16 // Holds the current opcode
	sp uint8  // Holds the current stack pointer value
	iv uint16 // Holds the index register value

	delayTimer uint8 // Timer used to count down a delay
	soundTimer uint8 // Timer used to count down to a beep being played

	shouldDraw bool   // Whether we should draw to the screen
	beeper     func() // Function to play a beep
}

// Initialise and return  an instance of the CHIP-8 VM
func Init() Chip8 {
	instance := Chip8{
		shouldDraw: true,
		pc:         0x200,
		beeper:     func() {},
	}

	// Load font into memory from bottom (we can use <0x200 for fonts)
	fmt.Printf("Loading fonts into memory\n")
	for i := 0; i < len(fontSet); i++ {
		log.Printf("\tLoading font [0x%X] into memory position [0x%X]\n", fontSet[i], instance.memory[i])
		instance.memory[i] = fontSet[i]
	}

	return instance
}

func (cpu *Chip8) GetBuffer() [32][64]uint8 {
	return cpu.display
}

func (cpu *Chip8) Draw() bool {
	sd := cpu.shouldDraw
	cpu.shouldDraw = false
	return sd
}

func (cpu *Chip8) AddBeep(fn func()) {
	cpu.beeper = fn
}

func (cpu *Chip8) Key(num uint8, down bool) {
	if down {
		cpu.key[num] = 1
	} else {
		cpu.key[num] = 0
	}
}

// Simulates a cycle of CHIP-8 CPU
//
//	 -> Read OpCode
//		-> Carry out instruction
//	 -> Increment program counter
//			-> PC is incremented by 2 as we operate on 2 opcodes at a time
//	 -> Check delay/sound timers
func (cpu *Chip8) Cycle() {
	cpu.oc = (uint16(cpu.memory[cpu.pc])<<8 | uint16(cpu.memory[cpu.pc+1]))
	fmt.Printf("Opcode: 0x%X\n", cpu.oc)

	/*
		Get the most significant 4-bits of the instruction by bitwise AND'ing with 0xF000
		Example:
			0x7522 & 0xF000 = 0x7000
	*/
	switch cpu.oc & 0xF000 {
	case 0x000:
		switch cpu.oc & 0x000F {
		case 0x0000: // 0x00E0: CLS (Clear screen)
			for i := 0; i < len(cpu.display); i++ {
				for j := 0; j < len(cpu.display[i]); j++ {
					cpu.display[i][j] = 0x0
				}
			}
			cpu.shouldDraw = true
			cpu.pc += 2
		case 0x000E: // 0x00EE: RET (Returns from a subroutine)
			cpu.sp--
			cpu.pc = cpu.stack[cpu.sp] + 2
		default:
			fmt.Printf("Invalid opcode: 0x%X\n", cpu.oc)
		}
	case 0x1000: // 0x1NNN: JP addr (Jumps to address at NNN)
		cpu.pc = cpu.oc & 0x0FFF
	case 0x2000: // 0x2NNN: CALL addr (Calls subroutine at NNN)
		cpu.stack[cpu.sp] = cpu.pc // Store current PC address
		cpu.sp++                   // Increment the stack pointer
		cpu.pc = cpu.oc & 0x0FFF   // Jump to address NNN
	case 0x3000: // 0x3XKK: SE Vx, byte (Skip next instruction if Vx == kk)
		if uint16(cpu.vx[(cpu.oc&0x0F00)>>8]) == cpu.oc&0x00FF {
			cpu.pc += 4 // Skips next 2 bytes and increments as usual
		} else {
			cpu.pc += 2 // Increments as usual
		}
	case 0x4000: // 0x4XNN: SNE Vx, byte (Skip next instruction if Vx != kk)
		if uint16(cpu.vx[(cpu.oc&0x0F00)>>8]) != cpu.oc&0x00FF {
			cpu.pc += 4 // Skips next 2 bytes and increments as usual
		} else {
			cpu.pc += 2 // Increments as usual
		}
	case 0x5000: // 0x5XY0: SE Vx, Vy (Skip next instruction if Vx == Vy)
		if cpu.vx[(cpu.oc&0x0F00)>>8] == cpu.vx[(cpu.oc&0x00F0)>>4] {
			cpu.pc += 4
		} else {
			cpu.pc += 2
		}
	case 0x6000: // 0x6XKK: LD Vx, byte (Loads value KK into register VX)
		cpu.vx[(cpu.oc&0x0F00)>>8] = uint8(cpu.oc & 0x00FF)
		cpu.pc += 2
	case 0x7000: // Ox7XKK: ADD Vx, byte (Adds the value KK to value in register Vx)
		cpu.vx[(cpu.oc&0x0F00)>>8] += uint8(cpu.oc & 0x00FF)
		cpu.pc += 2
	case 0x8000:
		switch cpu.oc & 0x000F {
		case 0x0000: // Ox8XY0: LD Vx, Vy (Loads value of Vy in register Vx)
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x0001: // 0x8XY1: OR Vx, Vy (Performs bitwise OR on Vx and Vy)
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.vx[(cpu.oc&0x0F00)>>8] | cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x0002: // 0x8XY2: AND Vx, Vy (Performs bitwise AND on Vx and Vy)
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.vx[(cpu.oc&0x0F00)>>8] & cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x0003: // 0x8XY3: XOR Vx, Vy (Performs bitwise XOR on Vx and Vy)
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.vx[(cpu.oc&0x0F00)>>8] ^ cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x0004: // 0x8XY4: ADD Vx, Vy (Add Vx and Vy, carry bit set if result > 255)
			//
			if cpu.vx[(cpu.oc&0x00F0)>>4] > (0xFF - cpu.vx[(cpu.oc&0x0F00)>>8]) {
				cpu.vx[0xF] = 1
			} else {
				cpu.vx[0xF] = 0
			}
			cpu.vx[(cpu.oc&0x0F00)>>8] += cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x0005: // 0x8XY5: SUB Vx, Vy (Vy subbed from Vx, if Vx > Vy then VF = 1)
			if cpu.vx[(cpu.oc&0x0F00)>>8] > cpu.vx[(cpu.oc&0x00F0)>>4] {
				cpu.vx[0xF] = 1
			} else {
				cpu.vx[0xF] = 0
			}
			cpu.vx[(cpu.oc&0x0F00)>>8] -= cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x0006: // 0x8XY6: SHR Vx {, Vy} (If lsb of Vx is 1, VF = 1, otherwise 0. Vx is divided by 2)
			cpu.vx[0xF] = cpu.vx[(cpu.oc&0x0F00)>>8] & 0x1
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.vx[(cpu.oc&0x0F00)>>8] >> 1
			cpu.pc += 2
		case 0x0007: // 0x8XY7: SUBN Vx, Vy (Set Vx = Vy - Vx, set VF = NOT borrow)
			if cpu.vx[(cpu.oc&0x00F0)>>4] > cpu.vx[(cpu.oc&0x0F00)>>8] {
				cpu.vx[0xF] = 1
			} else {
				cpu.vx[0xF] = 0
			}
			cpu.vx[(cpu.oc&0x0F00)>>8] -= cpu.vx[(cpu.oc&0x00F0)>>4]
			cpu.pc += 2
		case 0x000E: // 0x8XYE: SHL Vx, {, Vy} (If MSB of Vx is 1, VF = 1, otherwise 0, Vx doubled)
			cpu.vx[0xF] = cpu.vx[(cpu.oc&0x0F00)>>8] >> 7
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.vx[(cpu.oc&0x0F00)>>8] << 1
			cpu.pc += 2
		default:
			fmt.Printf("Invalid opcode: 0x%X\n", cpu.oc)
		}
	case 0x9000: // 0x9XY0: SNE Vx, Vy (Skips next instruction if Vx == Vy)
		if cpu.vx[(cpu.oc&0x0F00)>>8] != cpu.vx[(cpu.oc&0x00F0)>>4] {
			cpu.pc += 4
		} else {
			cpu.pc += 2
		}
	case 0xA000: // 0xANNN: LD I, addr (Load value at addr into register I)
		cpu.iv = cpu.oc & 0x0FFF
		cpu.pc += 2
	case 0xB000: // 0xBNNN: JP V0, addr (Jump to location NNN + V0)
		cpu.pc = (cpu.oc & 0x0FFF) + uint16(cpu.vx[0x0])
	case 0xC000: // 0xCXKK: RND Vx, byte (Set Vx to random byte AND KK)
		cpu.vx[(cpu.oc&0x0F00)>>8] = uint8(rand.Intn(256)) & uint8(cpu.oc&0x00FF)
		cpu.pc += 2
	case 0xD000: // 0xDXYN: DRW Vx, Vy, nibble (Dispay n-byte sprite starting at I at (Vx, Vy), set VF = collision)
		x := cpu.vx[(cpu.oc&0x0F00)>>8]
		y := cpu.vx[(cpu.oc&0x00F0)>>4]
		h := cpu.oc & 0x000F
		cpu.vx[0xF] = 0
		// Declaring i and j here as they must be uint16
		var j uint16 = 0
		var i uint16 = 0
		for i = 0; i < h; i++ {
			pixel := cpu.memory[cpu.iv+i]
			for j = 0; j < 8; j++ {
				if (pixel & (0x80 >> j)) != 0 {
					if cpu.display[(y + uint8(i))][x+uint8(j)] == 1 {
						cpu.vx[0xF] = 1
					}
					cpu.display[(y + uint8(i))][x+uint8(j)] ^= 1
				}
			}
		}
		cpu.shouldDraw = true
		cpu.pc += 2
	case 0xE000:
		switch cpu.oc & 0x00FF {
		case 0x009E: // 0xEX9E: SKP Vx (Skip next instruction if key with the value of Vx is pressed)
			if cpu.key[cpu.vx[(cpu.oc&0x0F00)>>8]] == 1 {
				cpu.pc += 4
			} else {
				cpu.pc += 2
			}
		case 0x00A1: // 0xEXA1: SKNP Vx (Skip next instruction if key with the value of Vx is not pressed)
			if cpu.key[cpu.vx[(cpu.oc&0x0F00)>>8]] == 0 {
				cpu.pc += 4
			} else {
				cpu.pc += 2
			}
		default:
			fmt.Printf("Invalid opcode: 0x%X\n", cpu.oc)
		}
	case 0xF000:
		switch cpu.oc & 0x00FF {
		case 0x0007: // 0xFX07: LD Vx, DT (Loads the value of the delay timer into Vx)
			cpu.vx[(cpu.oc&0x0F00)>>8] = cpu.delayTimer
			cpu.pc += 2
		case 0x000A: // 0xFX0A: LD Vx, K (Loads the value of key when pressed into Vx)
			pressed := false
			for i := 0; i < len(cpu.key); i++ {
				if cpu.key[i] != 0 {
					cpu.vx[(cpu.oc&0x0F00)>>8] = uint8(i)
					pressed = true
				}
			}
			if !pressed {
				return
			}
			cpu.pc += 2
		case 0x0015: // 0xFX15: LD DT, Vx (Set delayTimer to value at Vx)
			cpu.delayTimer = cpu.vx[(cpu.oc&0x0F00)>>8]
			cpu.pc += 2
		case 0x0018: // 0xFX18: LD ST, Vx (Set soundTimer to value at Vx)
			cpu.soundTimer = cpu.vx[(cpu.oc&0x0F00)>>8]
			cpu.pc += 2
		case 0x001E: // 0xFX1E: ADD I, Vx (Add value at Vx to I register)
			if cpu.iv+uint16(cpu.vx[(cpu.oc&0x0F00)>>8]) > 0xFFF {
				cpu.vx[0xF] = 1
			} else {
				cpu.vx[0xF] = 0
			}
			cpu.iv += uint16(cpu.vx[(cpu.oc&0x0F00)>>8])
			cpu.pc += 2
		case 0x0029: // 0xFX29: LD F, Vx (Set I register to location of sprite for digit Vx)
			cpu.iv = uint16(cpu.vx[(cpu.oc&0x0F00)>>8]) * 0x5
			cpu.pc += 2
		case 0x0033: // 0xFX33: LD B, Vx (Store BCD representation of Vx in mem locs I, I+1, I+2)
			cpu.memory[cpu.iv] = cpu.vx[(cpu.oc&0x0F00)>>8] / 100
			cpu.memory[cpu.iv+1] = cpu.vx[(cpu.oc&0x0F00)>>8] % 10
			cpu.memory[cpu.iv+2] = cpu.vx[(cpu.oc&0x0F00)>>8] / 10
			cpu.pc += 2
		case 0x0055: // 0xFX55: LD [I], Vx (Store registers V0 through Vx from mem starting at I)
			for i := 0; i < int((cpu.oc&0x0F00)>>8)+1; i++ {
				cpu.memory[uint16(i)+cpu.iv] = cpu.vx[i]
			}
			cpu.iv = ((cpu.oc & 0x0F00) >> 8) + 1
			cpu.pc += 2
		case 0x0065: // 0xFX65: LD Vx, [I] (Read registers V0 through Vx from mem starting at I)
			for i := 0; i < int((cpu.oc&0x0F00)>>8)+1; i++ {
				cpu.vx[i] = cpu.memory[cpu.iv+uint16(i)]
			}
			cpu.iv = ((cpu.oc & 0x0F00) >> 8) + 1
			cpu.pc += 2
		default:
			fmt.Printf("Invalid opcode: 0x%X\n", cpu.oc)
		}
	default:
		fmt.Printf("Invalid opcode: 0x%X\n", cpu.oc)
	}

	if cpu.delayTimer > 0 {
		cpu.delayTimer--
	}
	if cpu.soundTimer > 0 {
		if cpu.soundTimer == 1 {
			cpu.beeper()
		}
		cpu.soundTimer--
	}
}

func (cpu *Chip8) LoadProgram(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0777)
	if err != nil {
		return fmt.Errorf("Failed to read ROM file %s: %v", fileName, err)
	}
	defer file.Close()

	fStat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Failed to stat file %s: %v", fileName, err)
	}

	if int64(len(cpu.memory)-512) < fStat.Size() {
		// Program is too large
		return fmt.Errorf("ROM size (%v) is larger than available memory (%v)", fStat.Size(), (len(cpu.memory) - 512))
	}

	buffer := make([]byte, fStat.Size())
	if _, err := file.Read(buffer); err != nil {
		return err
	}

	for i := 0; i < len(buffer); i++ {
		cpu.memory[i+0x200] = buffer[i]
	}

	return nil
}
