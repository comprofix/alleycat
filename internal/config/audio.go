package config

// Beeper synth (DESIGN §7).
const (
	AudioSampleRate = 44100
	AudioBufferMs   = 50 // output buffer; keeps effects close to their events

	BeeperAmplitude = 0.3  // peak level at full volume, out of 1
	BeeperSlewMs    = 2.0  // attack and release, so notes start and stop without clicks
	SoftenCutoffHz  = 3000 // "Soften beeper" one-pole low-pass
	DefaultVolume   = 0.8  // out of 1
)
