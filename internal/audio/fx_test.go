package audio

import (
	"math"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/comprofix/alleycat/internal/config"
)

func newTestFX() *FX { return NewFX(rand.New(rand.NewPCG(3, 4))) }

func total(n []Note) time.Duration {
	var d time.Duration
	for _, x := range n {
		d += x.Dur
	}
	return d
}

func nearTicks(d time.Duration, ticks float64) bool {
	return math.Abs(d.Seconds()-Ticks(ticks).Seconds()) < 0.002
}

// Each effect lasts as long as the original's (DESIGN §7.3).
func TestFXLengths(t *testing.T) {
	fx := newTestFX()
	for _, c := range []struct {
		name  string
		notes []Note
		ticks float64
	}{
		{"mouse caught", fx.MouseCaught(), 1.5},
		{"squeak", fx.Squeak(), config.SqueakTicks},
		{"BONK", fx.Bonk(), config.BonkTicks},
		{"fence drop", fx.FenceDrop(), config.FenceDropTicks},
	} {
		if d := total(c.notes); !nearTicks(d, c.ticks) {
			t.Errorf("%s lasts %v, want %v ticks (%v)", c.name, d, c.ticks, Ticks(c.ticks))
		}
	}
	if d := total(fx.TrashCat()); math.Abs(d.Seconds()-0.06) > 0.001 {
		t.Errorf("trash cat lasts %v, want 3 notes of about 20 ms", d)
	}
}

func TestFXShapes(t *testing.T) {
	fx := newTestFX()
	if n := fx.MouseCaught(); n[1].Hz <= n[0].Hz {
		t.Errorf("mouse caught %v: want a blip up", n)
	}
	if n := fx.TrashCat(); !(n[0].Hz > n[1].Hz && n[1].Hz > n[2].Hz) {
		t.Errorf("trash cat %v: want three descending notes", n)
	}
	for _, n := range fx.Squeak() {
		if n.Hz < 1900 || n.Hz > 2330 || n.Dur != Ticks(1) {
			t.Fatalf("squeak note %+v: want a tick at 1.9–2.3 kHz", n)
		}
	}
	b := fx.Bonk()
	early, late := 0.0, 0.0
	for i := range 100 {
		early += b[i].Hz
		late += b[len(b)-1-i].Hz
	}
	if late >= early {
		t.Error("BONK should slide down")
	}
	// The fence drop is silent for 2 ticks, then crackles: a speaker push,
	// never a tone.
	f := fx.FenceDrop()
	if f[0].Hz != 0 || f[0].Dur != Ticks(2) {
		t.Errorf("fence drop starts %+v, want 2 silent ticks", f[0])
	}
	for _, n := range f[1:] {
		if n.Hz != 0 && n.Hz != 1 {
			t.Fatalf("fence drop note %+v: want pushes and rests only", n)
		}
	}
	// The dog fight is only the stepped sweep, one step a tick, rising and
	// wrapping round, for longer than any fight: no shrill warble.
	d := fx.DogFight()
	if total(d) < 30*time.Second {
		t.Errorf("fight sound lasts %v, want longer than any fight", total(d))
	}
	if d[1].Hz <= d[0].Hz {
		t.Error("the fight sweep should rise")
	}
	wraps := 0
	for i, n := range d {
		if n.Dur != Ticks(1) || n.Hz < 1700 || n.Hz > 4600 {
			t.Fatalf("fight step %d = %+v: want one tick between 1.7 and 4.6 kHz", i, n)
		}
		if i > 0 && n.Hz < d[i-1].Hz {
			wraps++
		}
	}
	if wraps < 100 {
		t.Errorf("the sweep wrapped %d times, want it to keep coming round", wraps)
	}
}

func TestNilFX(t *testing.T) {
	var fx *FX
	for _, n := range [][]Note{fx.MouseCaught(), fx.Squeak(), fx.Bonk(), fx.TrashCat(),
		fx.CanOn(), fx.CanOff(), fx.FenceDrop(), fx.DogFight()} {
		if n != nil {
			t.Fatal("a nil FX made notes")
		}
	}
}

// The room and heart-room effects last as long as the original's and move
// the right way (DESIGN §7.3).
func TestRoomFX(t *testing.T) {
	fx := newTestFX()
	for _, c := range []struct {
		name  string
		notes []Note
		ticks float64
	}{
		{"ZAP", fx.Zap(), config.ZapTicks},
		{"BYTE", fx.Byte(), config.ByteTicks},
		{"dog wake", fx.DogWake(), 2},
		{"reject fall", fx.RejectFall(14), 14},
		{"Hi?", fx.Hi(18), 18},
	} {
		if d := total(c.notes); math.Abs(d.Seconds()-Ticks(c.ticks).Seconds()) > 0.004 {
			t.Errorf("%s lasts %v, want %v ticks (%v)", c.name, d, c.ticks, Ticks(c.ticks))
		}
	}
	if d := total(fx.Smooch()); math.Abs(d.Seconds()-0.555) > 0.01 {
		t.Errorf("smooch lasts %v, want about 0.55 s", d)
	}

	for _, c := range []struct {
		name string
		n    []Note
		up   bool
	}{
		{"cage push", fx.CagePush(), true},
		{"fish eaten", fx.FishEaten(), true},
		{"fern eaten", fx.FernEaten(), true},
		{"bowl empty", fx.BowlEmpty(), true},
		{"gift pick", fx.GiftPick(), true},
		{"bird chirp", fx.BirdChirp(), false},
		{"rival bump", fx.RivalBump(), false},
		{"arrow hit", fx.ArrowHit(), false},
		{"gift drop", fx.GiftDrop(), false},
	} {
		if len(c.n) != 2 || (c.n[1].Hz > c.n[0].Hz) != c.up {
			t.Errorf("%s %v: want a two-tone blip, up %v", c.name, c.n, c.up)
		}
	}

	if hi, lo := fx.CageFall(134)[0].Hz, fx.CageFall(164)[0].Hz; lo >= hi || math.Abs(lo-1819) > 2 {
		t.Errorf("cage fall %.0f Hz at the top, %.0f at the floor: want falling to about 1819", hi, lo)
	}
	if a, b := fx.AirWarning(0)[0].Hz, fx.AirWarning(3)[0].Hz; b >= a {
		t.Error("each air warning should be lower than the last")
	}
	r := fx.RejectFall(14)
	for i := 1; i < len(r); i++ {
		if r[i].Hz >= r[i-1].Hz {
			t.Fatalf("reject fall %v: want it falling every tick", r)
		}
	}
	h := fx.Hi(4)
	if h[0].Hz == h[1].Hz || h[0].Hz != h[2].Hz {
		t.Errorf("Hi? %v: want two tones alternating every tick", h)
	}
	// The buzzes are speaker pushes, never tones.
	for _, n := range append(fx.Zap(), append(fx.Byte(), fx.DogWake()...)...) {
		if n.Hz != 0 && n.Hz != 1 {
			t.Fatalf("buzz note %+v: want pushes and rests only", n)
		}
	}
}
