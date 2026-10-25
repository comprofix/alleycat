# Alley Cat

A faithful rewrite of the original **Alley Cat** video game, created by **Bill Williams** and published by **Synapse Software** in 1983. This remake follows the 1984 IBM PC version.

It is written from scratch in Go with [Ebitengine](https://ebitengine.org/). The rules, layouts, speeds and timings follow the original. Every sprite, tune, sound effect and line of code is new: no original assets or code are used. This project is not affiliated with or endorsed by the original authors or rights holders.

## The game

You are Freddy, an alley cat. Climb the fences, trash cans and clotheslines, and leap through the open windows of the apartment block. Each window leads to a room with a task:

- **Birdcage:** knock the cage off its table and catch the bird.
- **Cheese:** dart through the holes in a giant cheese and catch the mice.
- **Fishbowl:** dive into the bowl and eat the fish, but watch your air and the eels.
- **Sleeping dogs:** drink the bowls dry without waking the dogs.
- **Bookshelf:** climb the bookcase and eat the three ferns on top, dodging the spider.

Clear enough rooms and Felicia appears in a window. Win her over in the heart room for a big bonus, and the next level begins.

The remake keeps the original's 320×200 screen and a PC-speaker-style sound made in code, with art in a 64-colour palette (Resurrect 64).

## Controls

| Key | Action |
|---|---|
| Arrow keys | Move, jump (Up) and drop (Down) |
| Alt | Special action: go through a cheese hole, eat or drink, put down a gift |
| K, H, T, A (or 1–4) | Choose a skill on the menu: Kitten, House Cat, Tomcat, Alley Cat |
| Esc | Paws mode (pause) |
| Ctrl+S | Sound on and off |
| Ctrl+R | Restart the game |
| Ctrl+M | Back to the skill menu |
| F9 | Square pixels or a 4:3 CRT shape |
| F11 or Alt+Enter | Fullscreen |
| F10 | Quit |

## Download

Builds for Windows and Linux (64-bit) are on the
[Releases](https://github.com/comprofix/alleycat/releases) page, with
`SHA256SUMS.txt` to check them against.

- **Windows:** unzip `alleycat-vX.Y.Z-windows-amd64.zip` and run
  `alleycat.exe`.
- **Linux:** extract `alleycat-vX.Y.Z-linux-amd64.tar.gz` and run
  `./alleycat`. It needs OpenGL, X11 and ALSA, which desktop systems
  already have.

Settings and the high score are saved in the folder the game is in, so keep
it somewhere you can write to.

## Building and running

You need [Go](https://go.dev/dl/) 1.27 or later.

```sh
go run ./cmd/alleycat                 # run the game
go build -o bin/ ./cmd/alleycat       # build for this computer
go test ./...                         # run the tests
```

**Windows** needs no C compiler. To build `alleycat.exe` from any system:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/alleycat.exe ./cmd/alleycat
```

**Linux** needs cgo and these development packages (Debian and Ubuntu names):

```sh
sudo apt install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev \
  libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
```

## Saved files

The game keeps two small files in the folder it runs from:

- `settings.json`: the screen shape, fullscreen, window size and sound;
- `scores.json`: the high score.

Delete either one to go back to the defaults.

## License

The code and the new art and music are released under the [MIT License](LICENSE). The *Alley Cat* name and the original game belong to their respective rights holders.
