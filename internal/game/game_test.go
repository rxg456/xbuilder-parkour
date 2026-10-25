package game

import (
	"math"
	"reflect"
	"testing"

	"github.com/rxg456/xbuilder-parkour/internal/player"
	"github.com/rxg456/xbuilder-parkour/internal/world"
)

func TestStateTransitions(t *testing.T) {
	g := New(42)
	if g.Player == nil || g.World == nil || g.State != Ready || g.Score != 0 || g.BestScore != 0 {
		t.Fatalf("unexpected initial game: %+v", g)
	}
	if g.Jump() {
		t.Fatal("jump accepted before start")
	}
	g.Start()
	if g.State != Running {
		t.Fatalf("Start state = %q, want %q", g.State, Running)
	}
	g.World.Distance = 120
	g.Start()
	if g.World.Distance != 120 {
		t.Fatal("repeated Start reset the run")
	}
	g.World.Obstacles = []world.Obstacle{{X: 170, Width: 20, Height: 20}}
	g.Update(1.0 / 120)
	if g.State != GameOver {
		t.Fatalf("collision state = %q, want %q", g.State, GameOver)
	}
	g.Start()
	if g.State != GameOver || g.Jump() {
		t.Fatal("Start or Jump bypassed GameOver")
	}
}

func TestInactiveStatesFreezeScene(t *testing.T) {
	for _, state := range []State{Ready, GameOver} {
		t.Run(string(state), func(t *testing.T) {
			g := New(42)
			g.State = state
			g.Player.Jump()
			g.World.Distance = 90
			g.Score = 9
			g.BestScore = 20
			assertUnchanged(t, g, 0.5)
		})
	}
}

func TestJumpAndLandingCollision(t *testing.T) {
	g := New(42)
	g.World.Obstacles = nil
	g.Start()
	if !g.Jump() || g.Player.Grounded || g.Jump() {
		t.Fatal("expected one grounded jump and rejection of a second jump")
	}
	g.Update(0.15)
	if g.Player.Y <= 20 || g.State != Running {
		t.Fatalf("player did not rise above obstacle height: %+v", g.Player)
	}
	// Keep the obstacle under the player while the world scrolls during descent.
	g.World.Obstacles = []world.Obstacle{{X: 170, Width: 300, Height: 20}}
	g.Update(1.0 / 120)
	if g.State != Running {
		t.Fatal("airborne player collided with a lower obstacle")
	}
	for i := 0; i < 120 && g.State == Running; i++ {
		g.Update(1.0 / 120)
	}
	if g.State != GameOver {
		t.Fatal("descending player did not collide")
	}
}

func TestCollisionUsesPlayerRectangle(t *testing.T) {
	for _, tc := range []struct {
		name     string
		obstacle world.Obstacle
		want     State
	}{
		{"left overlap", world.Obstacle{X: 150, Width: 20, Height: 20}, GameOver},
		{"right overlap", world.Obstacle{X: 199, Width: 20, Height: 20}, GameOver},
		{"behind", world.Obstacle{X: 100, Width: 20, Height: 20}, Running},
		{"ahead", world.Obstacle{X: 210, Width: 20, Height: 20}, Running},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := New(42)
			g.World.Obstacles = []world.Obstacle{tc.obstacle}
			g.Start()
			g.Update(1.0 / 120)
			if g.State != tc.want {
				t.Fatalf("state = %q, want %q", g.State, tc.want)
			}
		})
	}
}

func TestScoreAndBestScore(t *testing.T) {
	g := New(42)
	g.World.Obstacles = nil
	g.Start()
	for _, distance := range []float64{0, 9.9, 10, 99.9, 125} {
		g.World.Distance = distance
		// Advance without scrolling across the next score boundary.
		g.Update(0.000001)
		if want := int(distance / 10); g.Score != want {
			t.Fatalf("distance %v: score = %d, want %d", distance, g.Score, want)
		}
		if g.BestScore != 0 {
			t.Fatal("best score changed before game over")
		}
	}
	g.World.Obstacles = []world.Obstacle{{X: 170, Width: 20, Height: 20}}
	g.Update(1.0 / 120)
	if g.Score != 12 || g.BestScore != 12 || g.State != GameOver {
		t.Fatalf("unexpected final score: %+v", g)
	}
	g.Restart()
	g.Start()
	g.World.Distance = 30
	g.World.Obstacles = []world.Obstacle{{X: 170, Width: 20, Height: 20}}
	g.Update(1.0 / 120)
	if g.Score != 3 || g.BestScore != 12 {
		t.Fatalf("lower run replaced best score: %+v", g)
	}
}

func TestRunningAdvancesScene(t *testing.T) {
	g := New(42)
	g.World.Obstacles = nil
	g.Start()
	g.Jump()
	g.Update(0.25)
	if g.State != Running || g.Player.Y <= 0 || g.World.Distance <= 0 {
		t.Fatalf("running scene did not advance: %+v", g)
	}
	if g.Score != int(g.World.Distance/10) {
		t.Fatalf("score %d does not match distance %v", g.Score, g.World.Distance)
	}
}

func TestCollisionStopsRemainingUpdate(t *testing.T) {
	g := New(42)
	g.World.Obstacles = []world.Obstacle{{X: 210, Width: 20, Height: 30}}
	g.Start()
	g.Update(1)
	if g.State != GameOver || g.World.Distance <= 0 || g.World.Distance >= 30 {
		t.Fatalf("update did not stop at the approaching obstacle: state=%q distance=%v",
			g.State, g.World.Distance)
	}
	if g.Score != int(g.World.Distance/10) || g.BestScore != g.Score {
		t.Fatalf("collision did not finalize distance-based scores: %+v", g)
	}
	assertUnchanged(t, g, 1)
}

func TestRestart(t *testing.T) {
	for _, state := range []State{Ready, Running, GameOver} {
		t.Run(string(state), func(t *testing.T) {
			g := New(42)
			p, w := g.Player, g.World
			g.Player.Jump()
			g.Player.Update(0.1)
			g.World.Update(0.1)
			g.State = state
			g.Score = 9
			g.BestScore = 23
			g.Restart()
			if g.Player != p || g.World != w {
				t.Fatal("Restart replaced shared player or world pointers")
			}
			if g.State != Ready || g.Score != 0 || g.BestScore != 23 {
				t.Fatalf("unexpected restarted game: %+v", g)
			}
			if !g.Player.Grounded || g.Player.Y != player.DefaultConfig().GroundY || g.Player.VelocityY != 0 {
				t.Fatalf("player was not reset: %+v", g.Player)
			}
			fresh := New(42)
			if g.World.Distance != fresh.World.Distance || g.World.Speed != fresh.World.Speed ||
				!reflect.DeepEqual(g.World.Obstacles, fresh.World.Obstacles) {
				t.Fatal("world was not reset with the original seed")
			}
			g.Start()
			if !g.Jump() {
				t.Fatal("could not jump after restarting and starting")
			}
		})
	}
}

func TestInvalidDelta(t *testing.T) {
	for _, dt := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		g := New(42)
		g.Start()
		g.Jump()
		assertUnchanged(t, g, dt)
	}
}

func TestLargeDeltaIsCapped(t *testing.T) {
	g, other := New(42), New(42)
	for _, run := range []*Game{g, other} {
		run.World.Obstacles = nil
		run.Start()
		run.Jump()
	}
	g.Update(math.MaxFloat64)
	other.Update(1)
	if *g.Player != *other.Player || g.State != other.State || g.Score != other.Score ||
		g.World.Distance != other.World.Distance {
		t.Fatal("large delta differs from a one-second update")
	}
}

func assertUnchanged(t *testing.T, g *Game, dt float64) {
	t.Helper()
	before := *g
	beforePlayer := *g.Player
	distance, speed := g.World.Distance, g.World.Speed
	obstacles := make([]world.Obstacle, len(g.World.Obstacles))
	copy(obstacles, g.World.Obstacles)
	g.Update(dt)
	if *g != before || *g.Player != beforePlayer || g.World.Distance != distance ||
		g.World.Speed != speed || len(g.World.Obstacles) != len(obstacles) {
		t.Fatal("inactive or invalid update changed the scene")
	}
	for i, obstacle := range obstacles {
		if g.World.Obstacles[i] != obstacle {
			t.Fatal("inactive or invalid update moved an obstacle")
		}
	}
}
