// Package player implements vertical movement in a world where Y points up.
package player

import "math"

type Config struct {
	GroundY   float64
	JumpSpeed float64
	Gravity   float64
}

func DefaultConfig() Config {
	return Config{GroundY: 0, JumpSpeed: 600, Gravity: 1800}
}

type Player struct {
	Y         float64
	VelocityY float64
	Grounded  bool
	config    Config
}

// New replaces non-finite or non-positive speed/gravity with defaults.
func New(config Config) *Player {
	defaults := DefaultConfig()
	if !finite(config.GroundY) {
		config.GroundY = defaults.GroundY
	}
	if !finite(config.JumpSpeed) || config.JumpSpeed <= 0 {
		config.JumpSpeed = defaults.JumpSpeed
	}
	if !finite(config.Gravity) || config.Gravity <= 0 {
		config.Gravity = defaults.Gravity
	}
	p := &Player{config: config}
	p.Reset()
	return p
}

// Jump starts a jump only when grounded.
func (p *Player) Jump() bool {
	if !p.Grounded {
		return false
	}
	p.VelocityY = p.config.JumpSpeed
	p.Grounded = false
	return true
}

// Update ignores invalid dt and processes at most one second per call.
func (p *Player) Update(dt float64) {
	if !finite(dt) || dt <= 0 {
		return
	}
	if !finite(p.Y) || !finite(p.VelocityY) {
		p.Reset()
		return
	}
	if p.Grounded {
		p.Reset()
		return
	}

	const maxDelta = 1.0
	const maxStep = 1.0 / 120.0
	dt = math.Min(dt, maxDelta)
	steps := int(math.Ceil(dt / maxStep))
	step := dt / float64(steps)
	for i := 0; i < steps; i++ {
		// Exact constant-acceleration integration within each bounded step.
		p.Y += p.VelocityY*step - 0.5*p.config.Gravity*step*step
		p.VelocityY -= p.config.Gravity * step
		if !finite(p.Y) || !finite(p.VelocityY) {
			p.Reset()
			return
		}
		if p.Y <= p.config.GroundY && p.VelocityY <= 0 {
			p.Reset()
			return
		}
	}
}

// Reset returns the player to its configured ground with no vertical speed.
func (p *Player) Reset() {
	p.Y = p.config.GroundY
	p.VelocityY = 0
	p.Grounded = true
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
