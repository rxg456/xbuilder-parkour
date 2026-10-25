package world

import (
	"math"
	"reflect"
	"testing"
)

func TestNewAndGeneration(t *testing.T) {
	w := New(42)
	if w.Distance != 0 || w.Speed != 300 || len(w.Obstacles) != 1 {
		t.Fatalf("unexpected initial world: %+v", w)
	}
	if w.Obstacles[0].X != 960 {
		t.Fatalf("first obstacle X = %v, want 960", w.Obstacles[0].X)
	}
	generated := 1
	lastWidth := w.Obstacles[0].Width
	for frame := 0; frame < 20000; frame++ {
		w.Update(1.0 / 60.0)
		if len(w.Obstacles) == 0 || len(w.Obstacles) > 4 || cap(w.Obstacles) > 8 {
			t.Fatalf("frame %d: unbounded/empty obstacles: len=%d cap=%d",
				frame, len(w.Obstacles), cap(w.Obstacles))
		}
		for i, obstacle := range w.Obstacles {
			if obstacle.Width < 30 || obstacle.Width > 60 ||
				obstacle.Height < 30 || obstacle.Height > 60 {
				t.Fatalf("out-of-range obstacle: %+v", obstacle)
			}
			if obstacle.X > 960 || obstacle.X+obstacle.Width <= 0 {
				t.Fatalf("obstacle outside active range: %+v", obstacle)
			}
			if i > 0 {
				previous := w.Obstacles[i-1]
				gap := obstacle.X - previous.X - previous.Width
				// Subtract player width to measure the collision-free interval.
				freeTime := (gap - 40) / 550
				if freeTime < 2.0*600/1800+0.35-1e-9 {
					t.Fatalf("unsafe spacing: gap=%v freeTime=%v", gap, freeTime)
				}
			}
		}
		rightmost := w.Obstacles[len(w.Obstacles)-1]
		if rightmost.Width != lastWidth {
			generated++
			lastWidth = rightmost.Width
		}
	}
	if generated < 100 {
		t.Fatalf("only %d obstacles generated", generated)
	}
}

func TestMovementAndSpeedCap(t *testing.T) {
	w := New(1)
	start := w.Obstacles[0]
	w.Update(0.5)
	assertNear(t, "speed", w.Speed, 305)
	assertNear(t, "distance", w.Distance, 151.25)
	assertNear(t, "obstacle X", w.Obstacles[0].X, start.X-151.25)

	w.Reset(1)
	for i := 0; i < 30; i++ {
		w.Update(1)
	}
	assertNear(t, "capped speed", w.Speed, 550)
	assertNear(t, "integrated distance", w.Distance, 13375)
	w.Update(1)
	assertNear(t, "speed after cap", w.Speed, 550)
	assertNear(t, "distance after cap", w.Distance, 13925)
}

func TestRecycle(t *testing.T) {
	w := New(10)
	w.Obstacles = []Obstacle{
		{X: -29, Width: 30, Height: 30},
		{X: 400, Width: 45, Height: 50},
	}
	w.Update(0.01)
	if len(w.Obstacles) != 1 || w.Obstacles[0].Width != 45 {
		t.Fatalf("offscreen obstacle was not recycled: %+v", w.Obstacles)
	}
	assertNear(t, "remaining obstacle X", w.Obstacles[0].X, 396.9995)

	w.Obstacles = []Obstacle{{X: -60, Width: 30, Height: 30}}
	w.Update(1.0 / 120)
	if len(w.Obstacles) != 1 || w.Obstacles[0].X != 960 {
		t.Fatalf("empty field was not replenished: %+v", w.Obstacles)
	}
}

func TestSpeedBasedSpacing(t *testing.T) {
	slow, fast := New(1), New(1)
	fast.Speed = 550
	for _, w := range []*World{slow, fast} {
		w.Obstacles = []Obstacle{{X: 0, Width: 30, Height: 30}}
		w.Update(1.0 / 120)
		if len(w.Obstacles) != 2 {
			t.Fatalf("next obstacle was not generated: %+v", w.Obstacles)
		}
	}
	slowGap := slow.Obstacles[1].X - slow.Obstacles[0].X - 30
	fastGap := fast.Obstacles[1].X - fast.Obstacles[0].X - 30
	assertNear(t, "minimum gap", slowGap, 550*(2.0*600/1800+0.35)+40)
	assertNear(t, "high-speed gap", fastGap, 550*1.5+40)
	if fastGap <= slowGap {
		t.Fatal("spacing did not increase with speed")
	}
}

func TestCollides(t *testing.T) {
	w := &World{Obstacles: []Obstacle{{X: 180, Width: 40, Height: 60}}}
	tests := []struct {
		name       string
		x, y, w, h float64
		want       bool
	}{
		{"ground overlap", 160, 0, 40, 64, true},
		{"inside", 190, 10, 10, 10, true},
		{"left touching", 140, 0, 40, 64, false},
		{"right touching", 220, 0, 40, 64, false},
		{"top touching", 160, 60, 40, 64, false},
		{"below touching", 160, -64, 40, 64, false},
		{"jump over", 160, 61, 40, 64, false},
		{"partial vertical overlap", 160, 59, 40, 64, true},
		{"below partial overlap", 160, -63, 40, 64, true},
		{"left separated", 0, 0, 40, 64, false},
		{"right separated", 300, 0, 40, 64, false},
		{"zero width", 190, 0, 0, 64, false},
		{"negative height", 190, 0, 40, -1, false},
		{"nan x", math.NaN(), 0, 40, 64, false},
		{"nan y", 160, math.NaN(), 40, 64, false},
		{"infinite width", 160, 0, math.Inf(1), 64, false},
		{"infinite height", 160, 0, 40, math.Inf(-1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := w.Collides(tt.x, tt.y, tt.w, tt.h); got != tt.want {
				t.Fatalf("Collides = %v, want %v", got, tt.want)
			}
		})
	}
	w.Obstacles = nil
	if w.Collides(160, 0, 40, 64) {
		t.Fatal("empty world collides")
	}
}

func TestDefaultJumpClearsObstacle(t *testing.T) {
	w := New(9)
	w.Obstacles = []Obstacle{{X: 260, Width: 50, Height: 60}}
	passed := false
	groundCollision := false
	for frame := 1; frame <= 40; frame++ {
		w.Update(1.0 / 60.0)
		elapsed := float64(frame) / 60
		y := math.Max(0, 600*elapsed-0.5*1800*elapsed*elapsed)
		if w.Collides(160, y, 40, 64) {
			t.Fatalf("default jump collided at t=%v, feet=%v", elapsed, y)
		}
		groundCollision = groundCollision || w.Collides(160, 0, 40, 64)
		passed = passed || w.Obstacles[0].X+50 < 160
	}
	if !passed || !groundCollision {
		t.Fatalf("jump did not exercise obstacle passage: passed=%v groundCollision=%v",
			passed, groundCollision)
	}
}

func TestResetAndSeedReproducibility(t *testing.T) {
	a, b := New(123), New(123)
	if reflect.DeepEqual(a.Obstacles, New(456).Obstacles) {
		t.Fatal("different seeds produced identical initial obstacles")
	}
	for frame := 0; frame < 1000; frame++ {
		dt := float64(frame%5+1) / 120
		a.Update(dt)
		b.Update(dt)
		assertSameWorld(t, a, b)
	}
	a.Distance = -100
	a.Speed = 500
	a.Obstacles = append(a.Obstacles, Obstacle{X: 1, Width: 1, Height: 1})
	a.Reset(456)
	fresh := New(456)
	assertSameWorld(t, a, fresh)
	for frame := 0; frame < 1000; frame++ {
		a.Update(1.0 / 60)
		fresh.Update(1.0 / 60)
		assertSameWorld(t, a, fresh)
	}
}

func TestInvalidAndLargeDelta(t *testing.T) {
	for _, dt := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		w, untouched := New(42), New(42)
		w.Update(dt)
		assertSameWorld(t, w, untouched)
		// Also ensure an ignored update did not consume random values.
		for i := 0; i < 5; i++ {
			w.Update(1)
			untouched.Update(1)
		}
		assertSameWorld(t, w, untouched)
	}
	large, capped := New(42), New(42)
	large.Update(math.MaxFloat64)
	capped.Update(1)
	assertSameWorld(t, large, capped)
}

func TestUpdateSubsteps(t *testing.T) {
	a, b := New(7), New(7)
	a.Update(1)
	for i := 0; i < 120; i++ {
		b.Update(1.0 / 120)
	}
	assertSameWorld(t, a, b)
}

func TestZeroWorldAndInvalidSpeed(t *testing.T) {
	var w World
	w.Update(1.0 / 120)
	fresh := New(0)
	fresh.Update(1.0 / 120)
	assertSameWorld(t, &w, fresh)
	for _, speed := range []float64{math.NaN(), math.Inf(1), -1, 1000} {
		w := New(1)
		w.Speed = speed
		w.Update(0.1)
		if !finite(w.Speed) || w.Speed < 300 || w.Speed > 550 || !finite(w.Distance) {
			t.Fatalf("invalid speed was not normalized: %+v", w)
		}
	}
}

func assertSameWorld(t *testing.T, a, b *World) {
	t.Helper()
	if a.Speed != b.Speed || a.Distance != b.Distance ||
		!reflect.DeepEqual(a.Obstacles, b.Obstacles) {
		t.Fatalf("worlds differ:\na=%+v\nb=%+v", a, b)
	}
}

func assertNear(t *testing.T, name string, got, want float64) {
	t.Helper()
	if !finite(got) || math.Abs(got-want) > 1e-7 {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
