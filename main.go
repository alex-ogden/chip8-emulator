package main

import (
	"fmt"
	"os"
	"strconv"

	sdl "github.com/veandco/go-sdl2/sdl"

	beeper "chip8-emulator/beeper"
	chip8 "chip8-emulator/chip8"
)

const (
	CHIP8_DISP_HEIGHT int32 = 32 // Will be multiplied by modifier
	CHIP8_DISP_WIDTH  int32 = 64 // Will be multiplied by modifier
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("At least a ROM file must be provided")
		os.Exit(1)
	}

	// Set default value for modifier
	var modifier int32 = 10
	var fileName string = ""

	if len(os.Args) == 2 {
		// User should have only provided a ROM file here
		fileName = os.Args[1]
	} else if len(os.Args) == 3 {
		// User provided both a ROM file and modifier
		strModifier := os.Args[1]
		fileName = os.Args[2]
		val, err := strconv.ParseInt(strModifier, 10, 32)
		if err != nil {
			fmt.Printf("Modifier (first arg) should be a valid number")
			os.Exit(1)
		}
		if val > 0 {
			modifier = int32(val)
		} else {
			fmt.Printf("Modifier (first arg) must be greater than 0")
		}
	} else {
		fmt.Printf("Invalid number of arguments, expected 2 or 3, got %d\n", len(os.Args))
		os.Exit(1)
	}

	// Initialise CHIP8 emulator
	c8 := chip8.Init()
	if err := c8.LoadProgram(fileName); err != nil {
		fmt.Printf("%v\n", err)
	}

	// Initialise SDL2
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Printf("%v\n", err)
	}
	defer sdl.Quit()

	// Create window
	window, err := sdl.CreateWindow("CHIP8 - "+fileName, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, CHIP8_DISP_WIDTH*modifier, CHIP8_DISP_HEIGHT*modifier, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	defer window.Destroy()

	// Create render surface
	canvas, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	defer canvas.Destroy()

	// Initialise Beeper
	beep, err := beeper.Init()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	defer beep.Close()

	c8.AddBeep(func() {
		beep.Play()
	})

	// Main program loop
	for {
		c8.Cycle()
		// Draw if required
		if c8.Draw() {
			// Clear screen
			canvas.SetDrawColor(255, 0, 0, 255)
			canvas.Clear()

			// Get display buffer and render
			vector := c8.GetBuffer()
			for i := 0; i < len(vector); i++ {
				for j := 0; j < len(vector[i]); j++ {
					if vector[i][j] != 0 {
						canvas.SetDrawColor(255, 255, 0, 255)
					} else {
						canvas.SetDrawColor(255, 0, 0, 255)
					}
					canvas.FillRect(&sdl.Rect{
						Y: int32(i) * modifier,
						X: int32(j) * modifier,
						W: modifier,
						H: modifier,
					})
				}
			}

			canvas.Present()
		}

		// Poll for quit and keyboard events
		/*
			KeyMap:

				CHIP8 KEY						KEYBOARD KEY
			1 | 2 | 3 | C   ->   1 | 2 | 3 | 4
			4 | 5 | 6 | D   ->   Q | W | E | R
			7 | 8 | 9 | E   ->   A | S | D | F
			A | 0 | B | F   ->   Z | X | C | V
		*/

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch et := event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if et.Type == sdl.KEYUP {
					fmt.Printf("KEYUP EVENT\n")
					switch et.Keysym.Sym {
					case sdl.K_1:
						c8.Key(0x1, false)
					case sdl.K_2:
						c8.Key(0x2, false)
					case sdl.K_3:
						c8.Key(0x3, false)
					case sdl.K_4:
						c8.Key(0xC, false)
					case sdl.K_q:
						c8.Key(0x4, false)
					case sdl.K_w:
						c8.Key(0x5, false)
					case sdl.K_e:
						c8.Key(0x6, false)
					case sdl.K_r:
						c8.Key(0xD, false)
					case sdl.K_a:
						c8.Key(0x7, false)
					case sdl.K_s:
						c8.Key(0x8, false)
					case sdl.K_d:
						c8.Key(0x9, false)
					case sdl.K_f:
						c8.Key(0xE, false)
					case sdl.K_z:
						c8.Key(0xA, false)
					case sdl.K_x:
						c8.Key(0x0, false)
					case sdl.K_c:
						c8.Key(0xB, false)
					case sdl.K_v:
						c8.Key(0xF, false)
					}
				} else if et.Type == sdl.KEYDOWN {
					fmt.Printf("KEYDOWN EVENT\n")
					switch et.Keysym.Sym {
					case sdl.K_1:
						c8.Key(0x1, true)
					case sdl.K_2:
						c8.Key(0x2, true)
					case sdl.K_3:
						c8.Key(0x3, true)
					case sdl.K_4:
						c8.Key(0xC, true)
					case sdl.K_q:
						c8.Key(0x4, true)
					case sdl.K_w:
						c8.Key(0x5, true)
					case sdl.K_e:
						c8.Key(0x6, true)
					case sdl.K_r:
						c8.Key(0xD, true)
					case sdl.K_a:
						c8.Key(0x7, true)
					case sdl.K_s:
						c8.Key(0x8, true)
					case sdl.K_d:
						c8.Key(0x9, true)
					case sdl.K_f:
						c8.Key(0xE, true)
					case sdl.K_z:
						c8.Key(0xA, true)
					case sdl.K_x:
						c8.Key(0x0, true)
					case sdl.K_c:
						c8.Key(0xB, true)
					case sdl.K_v:
						c8.Key(0xF, true)
					}
				}
			}
		}

		// CHIP8 cpu clock should be set to 60Hz
		sdl.Delay(1000 / 60)
	}
}
