package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cvcvka5/2048-solver/internalgame"
)

const (
	SaveFile = "savegame.json"
	AIDepth  = 8
)

func main() {
	game := loadGame()

	for {
		printGrid(game)

		// Check if the game is already over before calculating moves
		state := game.CheckState()
		if state == internalgame.StateWon {
			fmt.Println("🎉 You reached 2048! You win!")
			break
		} else if state == internalgame.StateLost {
			fmt.Println("💀 Game Over! No moves left.")
			saveGameState(game) // Save the final state for post-mortem
			break
		}

		// AI Think Tank
		dir, err := game.CalculateBestMove(AIDepth)
		if err != nil {
			log.Fatalf("AI logic error: %v", err)
		}

		moved, _ := game.Shift(dir)
		if moved {
			game.SpawnTile()
			saveGameState(game) // Auto-save after every successful move
		}

		// Slow it down so we can see the magic happen
		time.Sleep(100 * time.Millisecond)
	}
}

// loadGame attempts to resume from JSON, otherwise starts fresh
func loadGame() *internalgame.Game {
	game := &internalgame.Game{}

	data, err := os.ReadFile(SaveFile)
	if err == nil {
		fmt.Println("Resuming existing game from save file...")
		if err := json.Unmarshal(data, game); err == nil {
			return game
		}
	}

	fmt.Println("Starting new game...")
	game.SpawnTile()
	game.SpawnTile()
	return game
}

// saveGameState uses the custom MarshalJSON we wrote
func saveGameState(g *internalgame.Game) {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		log.Printf("Warning: Failed to marshal game: %v", err)
		return
	}
	_ = os.WriteFile(SaveFile, data, 0644)
}

func printGrid(g *internalgame.Game) {
	// \033[H\033[2J is the ANSI escape code to clear the screen
	fmt.Print("\033[H\033[2J")
	fmt.Println("--- 2048 GO AI ---")

	for y := 0; y < internalgame.GridHeight; y++ {
		for x := 0; x < internalgame.GridWidth; x++ {
			val := g.GetTile(x, y)
			if val == 0 {
				fmt.Printf("| %4s ", ".")
			} else {
				// Colorize the output (optional, but looks cool)
				fmt.Printf("| %4d ", val)
			}
		}
		fmt.Println("|")
	}
	fmt.Println("-----------------")
}
