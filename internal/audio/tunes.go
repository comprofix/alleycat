package audio

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// The game's tunes (DESIGN §7.4): our own compositions, timed like the
// original's. Each is written as text: a note name and its length in master
// ticks ("E5:2"), or a rest ("r:2"). Sharps are written "C#5", flats "Bb4".
// Lengths may be fractions, so repeated notes can be split by short rests.
const (
	// titleTheme is played once on the title screen, about 48 s: a jaunty
	// strut in C, an intro, A A B A B A, and a coda. A beat is 4 ticks.
	titleIntro = `
		C4:2 E4:2 G4:2 E4:2 C4:2 E4:2 G4:2 E4:2
		F4:2 A4:2 C5:2 A4:2 G4:2 B4:2 D5:2 B4:2
		C4:2 E4:2 G4:2 E4:2 C4:2 E4:2 G4:2 E4:2
		G4:2 A4:2 B4:2 D5:2 G5:4 r:4`
	titleA = `
		E5:2 G5:2 A5:4 G5:2 E5:2 C5:4
		D5:2 E5:2 G5:4 E5:8
		C5:2 D5:2 E5:2 G5:2 A5:4 C6:4
		B5:4 A5:4 G5:8
		A5:2 G5:2 E5:4 G5:2 E5:2 D5:4
		C5:2 D5:2 E5:4 D5:8
		E5:2 G5:2 A5:2 C6:2 B5:4 G5:4`
	titleAEnd1 = `C6:8 r:8`
	titleAEnd2 = `C6:4 G5:2 E5:2 C5:8`
	titleB     = `
		F5:4 A5:4 C6:4 A5:4
		G5:2 A5:2 G5:2 F5:2 E5:8
		D5:4 F5:4 A5:4 F5:4
		E5:2 F5:2 E5:2 D5:2 C5:8
		A4:2 C5:2 E5:4 A5:4 E5:4
		B4:2 D5:2 F5:4 B5:4 A5:4
		G5:2 F5:2 E5:2 D5:2 C5:2 D5:2 E5:2 F5:2
		G5:12 r:4`
	titleCoda = `
		C6:2 G5:2 E5:2 C5:2 G5:2 E5:2 C5:2 G4:2
		C5:4 E5:4 G5:4 C6:4
		C6:12 r:4`

	// roomLeft follows "OOPS!": 14 notes of 2 ticks, cheeky and unbothered.
	roomLeft = `E5:2 G5:2 C6:2 B5:2 A5:2 G5:2 E5:2 D5:2 C5:2 D5:2 E5:2 G5:2 C5:2 r:2`

	// failure follows a lost life and plays under "HERE KITTY": 13 notes of
	// 2 ticks falling in A minor, then silence.
	failure = `A5:2 G5:2 F5:2 E5:2 D5:2 E5:2 C5:2 B4:2 A4:2 G#4:2 A4:2 E4:2 A3:2 r:2`

	// kissTune loops through the kiss sequence: a lilting waltz in C, a
	// note every 2 ticks at its quickest, 134 ticks a pass.
	kissTune = `
		E5:4 G5:2 C6:6 B5:2 A5:2 G5:4
		F5:2 A5:2 G5:6 E5:4 C5:2
		D5:4 F5:2 A5:6 G5:2 F5:2 E5:4
		D5:2 E5:2 C5:8 r:4
		E5:4 G5:2 C6:6 D6:2 C6:2 B5:4
		A5:2 C6:2 B5:6 G5:4 E5:2
		F5:4 A5:2 G5:4 E5:2 D5:4 C5:2
		C5:14`

	// danceTune is the dancing-cats interlude: 40 ticks of a can-can-style
	// gallop, a note a tick, repeats split by short rests.
	danceTune = `
		G5:0.6 r:0.4 G5:0.6 r:0.4 A5:1 G5:1 C6:0.6 r:0.4 C6:0.6 r:0.4 B5:1 r:1
		A5:0.6 r:0.4 A5:0.6 r:0.4 B5:1 A5:1 D6:0.6 r:0.4 D6:0.6 r:0.4 C6:1 r:1
		E6:1 D6:1 C6:1 B5:1 A5:1 G5:1 F5:1 E5:1
		D5:1 G5:1 B5:1 D6:1 C6:2 G5:1 C6:1
		E6:1 C6:1 G5:1 E5:1 C5:2 r:2`
)

// titleTempo stretches the title theme: 1.25 makes a beat 5 ticks, not 4
// (owner: a little slower, 2026-10-15).
const titleTempo = 1.25

// Title is the title theme, played once.
func Title() []Note {
	n := tune(titleIntro, titleA, titleAEnd1, titleA, titleAEnd2, titleB,
		titleA, titleAEnd1, titleB, titleA, titleAEnd2, titleCoda)
	for i := range n {
		n[i].Dur = time.Duration(float64(n[i].Dur) * titleTempo)
	}
	return n
}

// RoomLeft is the tune after leaving a room without losing a life.
func RoomLeft() []Note { return tune(roomLeft) }

// Failure is the tune after a lost life, and under "HERE KITTY".
func Failure() []Note { return tune(failure) }

// Kiss is the kiss tune; play it looping.
func Kiss() []Note { return tune(kissTune) }

// Dance is the dancing-cats tune.
func Dance() []Note { return tune(danceTune) }

// wipeFigure is our own rising 4-note figure for screen changes.
var wipeFigure = [4]string{"G5", "B5", "D6", "G6"}

// Wipe is the screen-change figure, repeating for d: one note per step of
// the change, rising (DESIGN §7.4).
func Wipe(d time.Duration) []Note {
	const step = 25 * time.Millisecond
	var n []Note
	for i := 0; time.Duration(i)*step < d; i++ {
		n = append(n, Note{Hz: noteHz(wipeFigure[i%len(wipeFigure)]), Dur: step})
	}
	return n
}

// The menus' sounds, as the Game Skeleton's (DESIGN §7.4): a blip as the
// cursor moves or a value changes, and two notes rising as an item is
// chosen.
var (
	menuMove    = []Note{{Hz: noteHz("E6"), Dur: 50 * time.Millisecond}}
	menuConfirm = []Note{{Hz: noteHz("C6"), Dur: 50 * time.Millisecond}, {Hz: noteHz("G6"), Dur: 100 * time.Millisecond}}
)

// MenuMove is the cursor moving or a value changing.
func MenuMove() []Note { return menuMove }

// MenuConfirm is a menu item chosen.
func MenuConfirm() []Note { return menuConfirm }

// tallyScale is the rising series the tally's rows step up.
var tallyScale = []string{"C5", "D5", "E5", "F5", "G5", "A5", "B5", "C6", "D6", "E6", "F6", "G6", "A6", "B6"}

// TallyNote is the note row i plays as it lights, one row every rowTicks
// (DESIGN §7.4). The series wraps round after its last note.
func TallyNote(i int, rowTicks float64) []Note {
	return []Note{{Hz: noteHz(tallyScale[i%len(tallyScale)]), Dur: Ticks(rowTicks)}}
}

// TallyTrill alternates the last row's note and the next every tick, for
// ticks, while the tally screen flashes.
func TallyTrill(lastRow, ticks int) []Note {
	a := noteHz(tallyScale[lastRow%len(tallyScale)])
	b := noteHz(tallyScale[(lastRow+1)%len(tallyScale)])
	n := make([]Note, ticks)
	for i := range n {
		n[i] = Note{Hz: a, Dur: Ticks(1)}
		if i%2 == 1 {
			n[i].Hz = b
		}
	}
	return n
}

// tune parses text tunes, joined in order. The tables are written in this
// file, so a mistake in one is a programming error.
func tune(parts ...string) []Note {
	var n []Note
	for _, p := range parts {
		for _, tok := range strings.Fields(p) {
			name, ticks, ok := strings.Cut(tok, ":")
			t, err := strconv.ParseFloat(ticks, 64)
			if !ok || err != nil || t <= 0 {
				panic(fmt.Sprintf("audio: bad note %q", tok))
			}
			hz := 0.0
			if name != "r" {
				hz = noteHz(name)
			}
			n = append(n, Note{Hz: hz, Dur: Ticks(t)})
		}
	}
	return n
}

// noteHz is a note's equal-tempered pitch, A4 = 440 Hz. A name is a letter,
// an optional # or b, and an octave: "C5", "G#4", "Bb3".
func noteHz(name string) float64 {
	semis := map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}
	s, ok := semis[name[0]]
	rest := name[1:]
	if ok && len(rest) > 1 {
		switch rest[0] {
		case '#':
			s, rest = s+1, rest[1:]
		case 'b':
			s, rest = s-1, rest[1:]
		}
	}
	oct, err := strconv.Atoi(rest)
	if !ok || err != nil {
		panic(fmt.Sprintf("audio: bad note name %q", name))
	}
	midi := 12*(oct+1) + s
	return 440 * math.Pow(2, float64(midi-69)/12)
}
