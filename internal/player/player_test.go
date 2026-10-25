package player

import (
	"math"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	want := Config{GroundY: 0, JumpSpeed: 600, Gravity: 1800}
	if got := DefaultConfig(); got != want {
		t.Fatalf("DefaultConfig() = %+v, want %+v", got, want)
	}
}

func TestNew(t *testing.T) {
	for _, ground := range []float64{0, 75, -30} {
		config := DefaultConfig()
		config.GroundY = ground
		p := New(config)
		assertGrounded(t, p, ground)
		if p.config != config {
			t.Fatalf("config = %+v, want %+v", p.config, config)
		}
		p.Update(0.1)
		assertGrounded(t, p, ground)
	}
}

func TestJump(t *testing.T) {
	p := New(DefaultConfig())
	if !p.Jump() || p.Grounded || p.VelocityY != 600 || p.Y != 0 {
		t.Fatalf("unexpected jump state: %+v", p)
	}
	p.Update(0.1)
	if math.Abs(p.Y-51) > 1e-9 || math.Abs(p.VelocityY-420) > 1e-9 || p.Grounded {
		t.Fatalf("unexpected airborne state: %+v", p)
	}
}

func TestRejectAirborneJump(t *testing.T) {
	p := New(DefaultConfig())
	p.Jump()
	for _, dt := range []float64{0, 0.1, 0.3} {
		p.Update(dt)
		before := *p
		if p.Jump() {
			t.Fatal("airborne jump accepted")
		}
		if *p != before {
			t.Fatalf("rejected jump changed state: %+v, before %+v", p, before)
		}
	}
}

func TestLanding(t *testing.T) {
	config := DefaultConfig()
	config.GroundY = 25
	p := New(config)
	p.Jump()
	for i := 0; i < 120; i++ {
		p.Update(1.0 / 120)
		if p.Y < config.GroundY {
			t.Fatalf("player below ground: %+v", p)
		}
	}
	assertGrounded(t, p, config.GroundY)
	if !p.Jump() {
		t.Fatal("jump after landing rejected")
	}
}

func TestReset(t *testing.T) {
	config := Config{GroundY: -20, JumpSpeed: 300, Gravity: 900}
	p := New(config)
	p.Jump()
	p.Update(0.1)
	p.Reset()
	assertGrounded(t, p, config.GroundY)
	if p.config != config || !p.Jump() || p.VelocityY != config.JumpSpeed {
		t.Fatalf("reset failed to preserve configuration: %+v", p)
	}
}

func TestInvalidDelta(t *testing.T) {
	cases := []struct {
		name string
		dt   float64
	}{
		{"zero", 0},
		{"negative", -1},
		{"nan", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, airborne := range []bool{false, true} {
				p := New(DefaultConfig())
				if airborne {
					p.Jump()
					p.Update(0.1)
				}
				before := *p
				p.Update(tc.dt)
				if *p != before {
					t.Fatalf("invalid dt changed state: %+v, before %+v", p, before)
				}
			}
		})
	}
}

func TestLargeDelta(t *testing.T) {
	p := New(DefaultConfig())
	p.Jump()
	p.Update(math.MaxFloat64)
	assertGrounded(t, p, 0)
}

func TestDeltaIsCapped(t *testing.T) {
	config := Config{GroundY: 0, JumpSpeed: 600, Gravity: 10}
	p := New(config)
	q := New(config)
	p.Jump()
	q.Jump()
	p.Update(100)
	q.Update(1)
	if *p != *q || p.Grounded {
		t.Fatalf("large dt was not capped: %+v, one second %+v", p, q)
	}
}

func TestInvalidConfig(t *testing.T) {
	cases := []Config{
		{},
		{GroundY: math.NaN(), JumpSpeed: math.NaN(), Gravity: math.NaN()},
		{GroundY: math.Inf(1), JumpSpeed: math.Inf(1), Gravity: math.Inf(-1)},
		{GroundY: 0, JumpSpeed: -1, Gravity: -1},
	}
	for _, config := range cases {
		p := New(config)
		if p.config != DefaultConfig() {
			t.Fatalf("invalid config was not normalized: %+v", p.config)
		}
		assertGrounded(t, p, 0)
	}
}

func TestNonFiniteState(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, corruptY := range []bool{false, true} {
			p := New(DefaultConfig())
			p.Jump()
			if corruptY {
				p.Y = value
			} else {
				p.VelocityY = value
			}
			p.Update(0.1)
			assertGrounded(t, p, 0)
		}
	}
}

func assertGrounded(t *testing.T, p *Player, ground float64) {
	t.Helper()
	if p.Y != ground || p.VelocityY != 0 || !p.Grounded {
		t.Fatalf("expected grounded player at %v, got %+v", ground, p)
	}
}
