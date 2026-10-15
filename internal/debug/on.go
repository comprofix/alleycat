//go:build debug

// Package debug reports whether this is a debug build (`-tags debug`),
// which draws the collision boxes and a TPS/FPS readout (DESIGN §5).
package debug

// Enabled is true in debug builds.
const Enabled = true
