// Command simulate exercises the same game model without an engine installation.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/rxg456/xbuilder-parkour/internal/game"
)

func main() {
	seed := flag.Int64("seed", 42, "obstacle random seed")
	seconds := flag.Float64("seconds", 30, "simulation duration (0, 3600]")
	autojump := flag.Bool("autojump", true, "demonstration jump controller")
	flag.Parse()
	if math.IsNaN(*seconds) || math.IsInf(*seconds, 0) || *seconds <= 0 || *seconds > 3600 {
		fmt.Fprintln(os.Stderr, "seconds must be finite and within (0, 3600]")
		os.Exit(2)
	}

	g := game.New(*seed)
	g.Start()
	const dt = 1.0 / 120
	elapsed := 0.0
	jumps := 0
	for elapsed < *seconds && g.State == game.Running {
		if *autojump && g.Player.Grounded {
			for _, obstacle := range g.World.Obstacles {
				// Start ascent about 0.18 seconds before reaching the obstacle.
				distance := obstacle.X - (160 + 40)
				if distance >= 0 && distance <= g.World.Speed*0.18 {
					if g.Jump() {
						jumps++
					}
					break
				}
			}
		}
		step := math.Min(dt, *seconds-elapsed)
		g.Update(step)
		elapsed += step
	}
	fmt.Printf("seed=%d state=%s seconds=%.2f score=%d best=%d distance=%.2f speed=%.2f jumps=%d\n",
		*seed, g.State, elapsed, g.Score, g.BestScore, g.World.Distance, g.World.Speed, jumps)
}
