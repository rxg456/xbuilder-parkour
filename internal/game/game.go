// Package game coordinates player movement, the world, and run state.
package game

import (
	"math"

	"github.com/rxg456/xbuilder-parkour/internal/player"
	"github.com/rxg456/xbuilder-parkour/internal/world"
)

type State string

const (
	Ready    State = "ready"
	Running  State = "running"
	GameOver State = "game_over"
)

const (
	playerX      = 160.0
	playerWidth  = 40.0
	playerHeight = 64.0
	maxDelta     = 1.0
	maxStep      = 1.0 / 120.0
)

type Game struct {
	Player    *player.Player
	World     *world.World
	State     State
	Score     int
	BestScore int
	seed      int64
}

func New(seed int64) *Game {
	return &Game{
		Player: player.New(player.DefaultConfig()),
		World:  world.New(seed),
		State:  Ready,
		seed:   seed,
	}
}

func (g *Game) Start() {
	if g.State == Ready {
		g.State = Running
	}
}

func (g *Game) Jump() bool {
	return g.State == Running && g.Player.Jump()
}

// Update ignores invalid deltas and caps elapsed time like player.Update.
func (g *Game) Update(dt float64) {
	if g.State != Running || dt <= 0 || math.IsNaN(dt) || math.IsInf(dt, 0) {
		return
	}
	// Check the current position too, before a moving obstacle can pass it.
	if g.collides() {
		g.finish()
		return
	}
	dt = math.Min(dt, maxDelta)
	steps := int(math.Ceil(dt / maxStep))
	step := dt / float64(steps)
	for i := 0; i < steps; i++ {
		g.Player.Update(step)
		g.World.Update(step)
		g.Score = int(g.World.Distance / 10)
		if g.collides() {
			g.finish()
			return
		}
	}
}

// Restart resets the run with the original seed, preserving the best score.
func (g *Game) Restart() {
	g.Player.Reset()
	g.World.Reset(g.seed)
	g.State = Ready
	g.Score = 0
}

func (g *Game) collides() bool {
	return g.World.Collides(playerX, g.Player.Y, playerWidth, playerHeight)
}

func (g *Game) finish() {
	g.Score = int(g.World.Distance / 10)
	if g.Score > g.BestScore {
		g.BestScore = g.Score
	}
	g.State = GameOver
}
