// Package clock provides the master clock (DESIGN §3.8): the original's
// 18.2 Hz BIOS tick, driven from the 60 TPS game loop by a fixed-point
// accumulator, plus Freddy's sub-tick at twice that rate.
package clock

// The BIOS tick is the PC timer input (1,193,182 Hz) divided by 65,536,
// about 18.2065 Hz.
const (
	timerHz      = 1193182
	timerDivisor = 65536
)

// TickHz is the master tick rate in ticks per second.
const TickHz = float64(timerHz) / timerDivisor

// FreddyStepsPerTick is Freddy's update rate relative to the master tick
// (DESIGN §3.8, confirmed in R1).
const FreddyStepsPerTick = 2

// Clock converts game-loop frames into master ticks and Freddy steps.
type Clock struct {
	period    int64 // accumulator units per game-loop frame threshold
	tickAcc   int64
	freddyAcc int64
}

// New returns a clock for a game loop running at tps frames per second.
func New(tps int) *Clock {
	return &Clock{period: int64(timerDivisor) * int64(tps)}
}

// Advance moves the clock on by one game-loop frame and returns how many
// master ticks and Freddy steps fell within it. Both are spread evenly over
// time: at 60 TPS each is 0 or 1 per frame.
func (c *Clock) Advance() (ticks, freddySteps int) {
	c.tickAcc += timerHz
	for c.tickAcc >= c.period {
		c.tickAcc -= c.period
		ticks++
	}
	c.freddyAcc += timerHz * FreddyStepsPerTick
	for c.freddyAcc >= c.period {
		c.freddyAcc -= c.period
		freddySteps++
	}
	return ticks, freddySteps
}
