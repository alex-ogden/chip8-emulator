mod chip8;

use minifb::{Key, Window, WindowOptions};
use std::env;
use std::fs;
use std::time::{Duration, Instant};

// Display properties
// CHIP8 display is 64x32 but we provide a scaler to ensure the screen is big enough
// for modern systems (default: 10)
const DISPLAY_WIDTH: usize = 64;
const DISPLAY_HEIGHT: usize = 32;

fn main() {
    // Parse arugments
    let mut args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("ERROR: Invalid number of arguments!");
        eprintln!(
            "Usage: {} <ROM file> [speed] [resolution-scale] [--debug]",
            args[0]
        );
        eprintln!("Defaults:");
        eprintln!("\tspeed: 10");
        eprintln!("\tresolution scale: 10");
        eprintln!("--debug can be used anywhere in args to enable terminal-based debugger");
        std::process::exit(1);
    }

    // If the user passes --debug, detect and enable debug mode, then remove from args to ensure
    // other arguments are in the correct place when parsing
    let debug_enabled: bool = args.contains(&"--debug".to_string());
    args.retain(|val| val != "--debug");

    let rom_path = &args[1];
    let speed: u32 = if args.len() >= 3 {
        args[2].parse().unwrap_or_else(|_| {
            eprintln!("ERROR: Speed must be a number");
            std::process::exit(1);
        })
    } else {
        10      // Default value for speed (clock cycles per frame)
    };
    let res_scale: u32 = if args.len() >= 4 {
        args[3].parse().unwrap_or_else(|_| {
            eprintln!("ERROR: Resolution scale must be a number");
            std::process::exit(1);
        })
    } else {
        10      // Default value for resolution scale (default provides a 640x320 display)
    };

    // Load ROM file
    let rom =
        fs::read(rom_path).unwrap_or_else(|_| panic!("ERROR: Failed to load ROM: {}", rom_path));

    // Initialise a new CHIP8 instance
    let mut chip8 = chip8::Chip8::new();
    chip8.load_rom(&rom);

    // Create window
    let mut window = Window::new(
        &format!("CHIP-8 Emulator - {}", rom_path),
        DISPLAY_WIDTH * res_scale as usize,
        DISPLAY_HEIGHT * res_scale as usize,
        WindowOptions::default(),
    )
    .expect("ERROR: Failed to create window");

    window.set_target_fps(60);

    // Initialise display buffer
    let mut buffer: Vec<u32> = vec![0; DISPLAY_WIDTH * DISPLAY_HEIGHT];
    let mut last_timer_update = Instant::now();

    // Main program loop
    while window.is_open() && !window.is_key_down(Key::Escape) {
        // Run several cycles per frame
        for _ in 0..speed {
            if let Err(e) = chip8.cycle(debug_enabled) {
                eprintln!("ERROR: {}", e);
                return;
            }
        }

        // Update timers
        if last_timer_update.elapsed() >= Duration::from_millis(16) {
            if chip8.delay_timer > 0 {
                chip8.delay_timer -= 1;
            }
            if chip8.sound_timer > 0 {
                chip8.sound_timer -= 1;
            }
            last_timer_update = Instant::now();
        }

        // Update input
        update_keys(&window, &mut chip8);

        // Build up the buffer
        for (i, row) in chip8.display.iter().enumerate() {
            for (j, &pixel) in row.iter().enumerate() {
                let idx = i * DISPLAY_WIDTH + j;
                buffer[idx] = if pixel { 0xFFFFFF } else { 0x000000 };
            }
        }

        // Update the window with the buffer
        window
            .update_with_buffer(&buffer, DISPLAY_WIDTH, DISPLAY_HEIGHT)
            .expect("Failed to update window");
    }
}

fn update_keys(window: &Window, chip8: &mut chip8::Chip8) {
    chip8.keys[0x1] = window.is_key_down(Key::Key1);
    chip8.keys[0x2] = window.is_key_down(Key::Key2);
    chip8.keys[0x3] = window.is_key_down(Key::Key3);
    chip8.keys[0xC] = window.is_key_down(Key::Key4);

    chip8.keys[0x4] = window.is_key_down(Key::Q);
    chip8.keys[0x5] = window.is_key_down(Key::W);
    chip8.keys[0x6] = window.is_key_down(Key::E);
    chip8.keys[0xD] = window.is_key_down(Key::R);

    chip8.keys[0x7] = window.is_key_down(Key::A);
    chip8.keys[0x8] = window.is_key_down(Key::S);
    chip8.keys[0x9] = window.is_key_down(Key::D);
    chip8.keys[0xE] = window.is_key_down(Key::F);

    chip8.keys[0xA] = window.is_key_down(Key::Z);
    chip8.keys[0x0] = window.is_key_down(Key::X);
    chip8.keys[0xB] = window.is_key_down(Key::C);
    chip8.keys[0xF] = window.is_key_down(Key::V);
}
