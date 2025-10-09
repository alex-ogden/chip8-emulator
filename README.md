# chip8-emulator

This is a CHIP8 Emulator written (mostly) from scratch using Go.

### To-do:
- Implement my own SDL/OpenGL code for handling graphics (currently borrowing from [skatiyar/go-chip8](https://github.com/skatiyar/go-chip8/tree/master))
- Test with some dedicated CHIP8 test roms
- Re-implement in C or some other language at some point

### How to build

For most systems you should be fine with the following:

```sh
go build -o [binary name] -v
# Then run with
./[binary name] [modifier] path/to/rom
```

However for MacOS on Apple Silicon, you may have to do the following:

```sh
export CGO_ENABLED=1
export GOARCH=arm64
go build -o [binary name] -v
```

### How to run

Once you've compiled the binary, you should be able to run it as follows (this assumes you've built a binary called `chip8-emulator`):

```sh
./chip8-emulator [modifier] [path-to-rom]
```

The `modifier` is the amount you scale the original 64x32 resolution of the CHIP8's display to fit your display. Usually a modifier of 
`10` should work fine (this results in a 640x320 display) but you may wish to increase/decrease this to fit your display.

### Structure
```
├── beeper
│   └── beeper.go
├── build.sh
├── chip8
│   └── chip8.go
├── chip8-emulator
├── go.mod
├── go.sum
├── LICENSE
├── main.go
├── README.md
└── roms
    ├── chip8-picture.ch8
    ├── filter.ch8
    ├── invaders.c8
    └── pong.c8
```

- beeper - This is the code for handling the CHIP8's audio system.
- chip8 - This is where the CHIP8 implementation lives
- roms - These are some sample ROMs you can use with the CHIP8 emulator
