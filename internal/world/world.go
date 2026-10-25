// Package world implements a deterministic, horizontally scrolling obstacle field.
package world

import (
	"math"
	"math/rand"
)

const (
	visibleWidth = 960.0
	baseSpeed    = 300.0
	maxSpeed     = 550.0
	acceleration = 10.0
	playerWidth  = 40.0
	airTime      = 2.0 * 600.0 / 1800.0
	recoveryTime = 0.35
	maxDelta     = 1.0
	maxStep      = 1.0 / 120.0
)

// Obstacle rests on the ground; X is its left edge.
type Obstacle struct {
	X, Width, Height float64
}

// World owns its random source. It is not safe for concurrent use.
type World struct {
	Obstacles []Obstacle
	Distance  float64
	Speed     float64
	rng       *rand.Rand
}

// New starts at base speed with an obstacle at the right edge of the viewport.
func New(seed int64) *World {
	w := &World{}
	w.Reset(seed)
	return w
}

// Reset discards all progress and restarts the seeded obstacle sequence.
func (w *World) Reset(seed int64) {
	w.Distance = 0
	w.Speed = baseSpeed
	w.rng = rand.New(rand.NewSource(seed))
	w.Obstacles = []Obstacle{w.obstacle(visibleWidth)}
}

// Update ignores invalid dt and processes at most one second per call.
// Distance measures scrolling distance, and speed increases up to 550.
func (w *World) Update(dt float64) {
	if !finite(dt) || dt <= 0 {
		return
	}
	if w.rng == nil {
		w.Reset(0)
	}
	if !finite(w.Speed) {
		w.Speed = baseSpeed
	}
	w.Speed = math.Max(baseSpeed, math.Min(w.Speed, maxSpeed))

	dt = math.Min(dt, maxDelta)
	steps := int(math.Ceil(dt / maxStep))
	step := dt / float64(steps)
	for i := 0; i < steps; i++ {
		// Integrate acceleration exactly, including the speed-cap transition.
		accelTime := math.Min(step, (maxSpeed-w.Speed)/acceleration)
		travel := w.Speed*accelTime + 0.5*acceleration*accelTime*accelTime
		w.Speed = math.Min(maxSpeed, w.Speed+acceleration*accelTime)
		travel += w.Speed * (step - accelTime)
		w.Distance += travel

		kept := w.Obstacles[:0]
		for _, obstacle := range w.Obstacles {
			obstacle.X -= travel
			if obstacle.X+obstacle.Width > 0 {
				kept = append(kept, obstacle)
			}
		}
		// Clear removed values without retaining a growing obstacle history.
		clear(w.Obstacles[len(kept):])
		w.Obstacles = kept

		if len(w.Obstacles) == 0 {
			w.Obstacles = append(w.Obstacles, w.obstacle(visibleWidth))
		}
		last := w.Obstacles[len(w.Obstacles)-1]
		// The floor preserves air/recovery time even after later acceleration.
		gap := math.Max(w.Speed*1.5, maxSpeed*(airTime+recoveryTime)) + playerWidth
		nextX := last.X + last.Width + gap
		if nextX <= visibleWidth {
			w.Obstacles = append(w.Obstacles, w.obstacle(nextX))
		}
	}
}

// Collides tests strict AABB overlap. playerX is the left edge and playerY
// is the feet position, with Y increasing upward from ground level zero.
// Touching edges, invalid coordinates, and non-positive sizes do not collide.
func (w *World) Collides(playerX, playerY, playerWidth, playerHeight float64) bool {
	if !finite(playerX) || !finite(playerY) ||
		!finite(playerWidth) || !finite(playerHeight) ||
		playerWidth <= 0 || playerHeight <= 0 {
		return false
	}
	for _, obstacle := range w.Obstacles {
		if !finite(obstacle.X) || !finite(obstacle.Width) || !finite(obstacle.Height) ||
			obstacle.Width <= 0 || obstacle.Height <= 0 {
			continue
		}
		if playerX < obstacle.X+obstacle.Width &&
			playerX+playerWidth > obstacle.X &&
			playerY < obstacle.Height && playerY+playerHeight > 0 {
			return true
		}
	}
	return false
}

func (w *World) obstacle(x float64) Obstacle {
	return Obstacle{
		X:      x,
		Width:  30 + w.rng.Float64()*30,
		Height: 30 + w.rng.Float64()*30,
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
