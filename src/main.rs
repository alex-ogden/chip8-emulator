mod chip8;

use std::fs;
use std::env;
use std::time::{Duration, Instant};
use minifb::{Key, Window, WindowOptions};

// Constants
const DISPLAY_WIDTH: usize = 64;
const DISPLAY_HEIGHT: usize = 32;

fn main() {
    // Parse arugments
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("ERROR: Invalid number of arguments!");
        eprintln!("Usage: {} <ROM file> [speed] [resolution-scale]", args[0]);
        std::process::exit(1);
    }
    
    let rom_path = &args[1];
    let speed: u32 = if args.len() >= 3 {
        args[2].parse().unwrap_or_else(|_| {
            eprintln!("ERROR: Speed must be a number");
            std::process::exit(1);
        })
    } else {
        10
    };
    let res_scale: u32 = if args.len() >= 4 {
        args[3].parse().unwrap_or_else(|_| {
            eprintln!("ERROR: Resolution scale must be a number");
            std::process::exit(1);
        })
    } else {
        10
    };

    // Load ROM file
    let rom = fs::read(rom_path)
        .unwrap_or_else(|_| panic!("ERROR: Failed to load ROM: {}", rom_path));

    let mut chip8 = chip8::Chip8::new();
    chip8.load_rom(&rom);

    // Create window
    let mut window = Window::new(
        &format!("CHIP-8 Emulator - {}", rom_path),
        DISPLAY_WIDTH * res_scale as usize,
        DISPLAY_HEIGHT * res_scale as usize,
        WindowOptions::default(),
    ).expect("ERROR: Failed to create window");

    window.set_target_fps(60);
    //window.limit_update_rate(Some(Duration::from_micros(16600))); // ~60FPS

    let mut buffer: Vec<u32> = vec![0; DISPLAY_WIDTH * DISPLAY_HEIGHT];
    let mut last_timer_update = Instant::now();

    // Main program loop
    while window.is_open() && !window.is_key_down(Key::Escape) {
        // Run several cycles per frame
        for _ in 0..speed {
            if let Err(e) = chip8.cycle() {
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

        // Render display to buffer
        for (i, row) in chip8.display.iter().enumerate() {
            for (j, &pixel) in row.iter().enumerate() {
                let idx = i * DISPLAY_WIDTH + j;
                buffer[idx] = if pixel { 0xFFFFFF } else { 0x000000 };
            }
        }

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
