package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cvcvka5/2048-solver/internalgame"
)

func main() {
	// 1. Define CLI Flags
	gridInput := flag.String("grid", "", "JSON string of the grid (e.g. '{\"grid\":[[0,0,2,0],...]}')")
	depth := flag.Int("depth", 8, "Search depth for the AI")
	flag.Parse()

	// 2. Handle Logic: Single Eval vs. Continuous Play
	if *gridInput != "" {
		runSingleEvaluation(*gridInput, *depth)
		return
	}

	runAutoPlayer(*depth)
}

// Response matches what Python expects
type AIResponse struct {
	BestMove  string  `json:"best_move"`
	Direction int     `json:"direction"`
	Score     float64 `json:"heuristic_score"`
	Duration  string  `json:"duration"`
}

func runSingleEvaluation(rawJson string, depth int) {
	game := &internalgame.Game{}
	// Clean the input string in case of shell quoting issues
	rawJson = strings.Trim(rawJson, "'")

	if err := json.Unmarshal([]byte(rawJson), game); err != nil {
		// Output error as JSON so Python doesn't crash
		fmt.Printf(`{"error": "unmarshal failed: %v"}`+"\n", err)
		os.Exit(1)
	}

	start := time.Now()
	dir, _ := game.CalculateBestMove(depth)
	// Get the score for the state we just evaluated
	score := game.CalculateHeuristic()

	resp := AIResponse{
		BestMove:  moveToString(dir),
		Direction: int(dir),
		Score:     score,
		Duration:  time.Since(start).String(),
	}

	// Output ONLY the JSON to stdout
	finalJson, _ := json.Marshal(resp)
	fmt.Println(string(finalJson))
}

// runAutoPlayer is the loop that plays the game indefinitely
func runAutoPlayer(depth int) {
	game := &internalgame.Game{}
	game.SpawnTile()
	game.SpawnTile()

	for {
		printGrid(game)

		state := game.CheckState()
		if state == internalgame.StateWon {
			fmt.Println("🎉 2048 REACHED! AI WINS!")
			break
		} else if state == internalgame.StateLost {
			fmt.Println("💀 GAME OVER! No moves left.")
			break
		}

		dir, err := game.CalculateBestMove(depth)
		if err != nil {
			log.Fatal(err)
		}

		moved, _ := game.Shift(dir)
		if moved {
			game.SpawnTile()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// --- Helpers ---

func printGrid(g *internalgame.Game) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println("--- 2048 GO CLI ---")
	for y := 0; y < internalgame.GridHeight; y++ {
		for x := 0; x < internalgame.GridWidth; x++ {
			val := g.GetTile(x, y)
			if val == 0 {
				fmt.Printf("| %4s ", ".")
			} else {
				fmt.Printf("| %4d ", val)
			}
		}
		fmt.Println("|")
	}
	fmt.Println("-----------------")
}

func moveToString(d internalgame.ShiftDirection) string {
	return []string{"UP", "DOWN", "LEFT", "RIGHT"}[d]
}
