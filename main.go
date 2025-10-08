package main

import (
	"os"
	"strconv"

	sdl "github.com/veandco/go-sdl2/sdl"

	beeper "chip8-emulator/beeper"
	chip8 "chip8-emulator/chip8"
)

const (
	CHIP8_DISP_HEIGHT int32 = 32
	CHIP8_DISP_WIDTH  int32 = 64
)

func main() {
	if len(os.Args) < 3 {
		panic("Please provide modifier and c8 ROM file")
	}

	fileName := os.Args[2]
	var modifier int32 = 10

	if len(os.Args) == 3 {
		if val, err := strconv.ParseInt(os.Args[1], 10, 32); err != nil {
			panic(err)
		} else {
			if val > 10 {
				modifier = int32(val)
			}
		}
	}

	// Initialise CHIP8
	c8 := chip8.Init()
	if err := c8.LoadProgram(fileName); err != nil {
		panic(err)
	}

	// Initialise SDL2
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// Initialise Beeper
	beep, err := beeper.Init()
	if err != nil {
		panic(err)
	}
	defer beep.Close()

	c8.AddBeep(func() {
		beep.Play()
	})

	// Create window
	window, err := sdl.CreateWindow("CHIP8 - "+fileName, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, CHIP8_DISP_WIDTH*modifier, CHIP8_DISP_HEIGHT*modifier, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// Create render surface
	canvas, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		panic(err)
	}
	defer canvas.Destroy()

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
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch et := event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if et.Type == sdl.KEYUP {
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
