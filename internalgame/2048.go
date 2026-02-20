package internalgame

import (
	"encoding/json"
	"math"
	"math/rand"
)

// --- Types & Constants ---

type TileValue uint16
type GameState int
type ShiftDirection int

const (
	GridWidth  = 4
	GridHeight = 4
)

const (
	ShiftUp ShiftDirection = iota
	ShiftDown
	ShiftLeft
	ShiftRight
)

var ShiftDirections = [4]ShiftDirection{ShiftUp, ShiftDown, ShiftLeft, ShiftRight}

const (
	StateUndetermined GameState = iota
	StateWon
	StateLost
)

type Game struct {
	grid [GridHeight][GridWidth]TileValue
}

// SetGrid allows us to manually inject a state for testing
func (g *Game) SetGrid(newGrid [GridHeight][GridWidth]TileValue) {
	g.grid = newGrid
}

// GetGrid returns the current board
func (g *Game) GetGrid() [GridHeight][GridWidth]TileValue {
	return g.grid
}

// --- Core Game Logic ---

func shiftLine(line [GridWidth]TileValue) ([GridWidth]TileValue, bool) {
	var next [GridWidth]TileValue
	idx := 0
	for i := 0; i < GridWidth; i++ {
		if line[i] != 0 {
			next[idx] = line[i]
			idx++
		}
	}
	for i := 0; i < GridWidth-1; i++ {
		if next[i] != 0 && next[i] == next[i+1] {
			next[i] *= 2
			next[i+1] = 0
			i++
		}
	}
	var final [GridWidth]TileValue
	idx = 0
	for i := 0; i < GridWidth; i++ {
		if next[i] != 0 {
			final[idx] = next[i]
			idx++
		}
	}
	return final, final != line
}

func (g *Game) CanShift(d ShiftDirection) (bool, error) {
	switch d {
	case ShiftUp:
		for x := 0; x < GridWidth; x++ {
			for y := 1; y < GridHeight; y++ {
				if g.grid[y][x] != 0 && (g.grid[y-1][x] == 0 || g.grid[y-1][x] == g.grid[y][x]) {
					return true, nil
				}
			}
		}
	case ShiftDown:
		for x := 0; x < GridWidth; x++ {
			for y := 0; y < GridHeight-1; y++ {
				if g.grid[y][x] != 0 && (g.grid[y+1][x] == 0 || g.grid[y+1][x] == g.grid[y][x]) {
					return true, nil
				}
			}
		}
	case ShiftLeft:
		for y := 0; y < GridHeight; y++ {
			for x := 1; x < GridWidth; x++ {
				if g.grid[y][x] != 0 && (g.grid[y][x-1] == 0 || g.grid[y][x-1] == g.grid[y][x]) {
					return true, nil
				}
			}
		}
	case ShiftRight:
		for y := 0; y < GridHeight; y++ {
			for x := 0; x < GridWidth-1; x++ {
				if g.grid[y][x] != 0 && (g.grid[y][x+1] == 0 || g.grid[y][x+1] == g.grid[y][x]) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (g *Game) Shift(d ShiftDirection) (bool, error) {
	can, _ := g.CanShift(d)
	if !can {
		return false, nil
	}
	changed := false
	switch d {
	case ShiftUp:
		for x := 0; x < GridWidth; x++ {
			var col [GridHeight]TileValue
			for y := 0; y < GridHeight; y++ {
				col[y] = g.grid[y][x]
			}
			newCol, moved := shiftLine(col)
			if moved {
				changed = true
				for y := 0; y < GridHeight; y++ {
					g.grid[y][x] = newCol[y]
				}
			}
		}
	case ShiftDown:
		for x := 0; x < GridWidth; x++ {
			var col [GridHeight]TileValue
			for y := 0; y < GridHeight; y++ {
				col[y] = g.grid[GridHeight-1-y][x]
			}
			newCol, moved := shiftLine(col)
			if moved {
				changed = true
				for y := 0; y < GridHeight; y++ {
					g.grid[GridHeight-1-y][x] = newCol[y]
				}
			}
		}
	case ShiftLeft:
		for y := 0; y < GridHeight; y++ {
			newRow, moved := shiftLine(g.grid[y])
			if moved {
				changed = true
				g.grid[y] = newRow
			}
		}
	case ShiftRight:
		for y := 0; y < GridHeight; y++ {
			var row [GridWidth]TileValue
			for x := 0; x < GridWidth; x++ {
				row[x] = g.grid[y][GridWidth-1-x]
			}
			newRow, moved := shiftLine(row)
			if moved {
				changed = true
				for x := 0; x < GridWidth; x++ {
					g.grid[y][GridWidth-1-x] = newRow[x]
				}
			}
		}
	}
	return changed, nil
}

// --- AI & Heuristics ---

func (g Game) CalculateBestMove(depth int) (ShiftDirection, error) {
	var bestMove ShiftDirection
	maxScore := -math.MaxFloat64

	for _, dir := range ShiftDirections {
		temp := g
		if moved, _ := temp.Shift(dir); moved {
			temp.SpawnTile()
			// We use depth - 1 because we've already simulated the first step
			score := temp.evaluatePath(depth - 1)
			if score > maxScore {
				maxScore = score
				bestMove = dir
			}
		}
	}
	return bestMove, nil
}

func (g Game) evaluatePath(depth int) float64 {
	if depth <= 0 {
		return g.CalculateHeuristic()
	}

	bestScore := -math.MaxFloat64
	movedAny := false

	for _, dir := range ShiftDirections {
		temp := g
		if moved, _ := temp.Shift(dir); moved {
			movedAny = true
			temp.SpawnTile()
			score := temp.evaluatePath(depth - 1)
			if score > bestScore {
				bestScore = score
			}
		}
	}

	if !movedAny {
		return -1e9
	} // Dead end penalty
	return bestScore
}

func (g Game) CalculateHeuristic() float64 {
	if g.CheckState() == StateLost {
		return -1e12
	}

	var score, penalty, empty float64

	// Snake Weights favoring the Bottom-Left corner
	weights := [4][4]float64{
		{0, 1, 2, 3},
		{7, 6, 5, 4},
		{8, 9, 10, 11},
		{15, 14, 13, 12},
	}

	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			val := float64(g.grid[y][x])
			if val == 0 {
				empty++
				continue
			}

			logVal := math.Log2(val)
			score += logVal * weights[y][x]

			// Smoothness penalty: check neighbors
			if x < 3 {
				right := float64(g.grid[y][x+1])
				if right > 0 {
					penalty += math.Abs(logVal - math.Log2(right))
				}
			}
			if y < 3 {
				down := float64(g.grid[y+1][x])
				if down > 0 {
					penalty += math.Abs(logVal - math.Log2(down))
				}
			}
		}
	}

	return score - (penalty * 15.0) + (empty * 10.0)
}

// --- Helpers ---

func (g *Game) SpawnTile() {
	var empties []struct{ x, y int }
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			if g.grid[y][x] == 0 {
				empties = append(empties, struct{ x, y int }{x, y})
			}
		}
	}
	if len(empties) > 0 {
		p := empties[rand.Intn(len(empties))]
		g.grid[p.y][p.x] = 2
		if rand.Float64() < 0.1 {
			g.grid[p.y][p.x] = 4
		}
	}
}

func (g *Game) CheckState() GameState {
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			if g.grid[y][x] >= 2048 {
				return StateWon
			}
		}
	}
	for _, dir := range ShiftDirections {
		if can, _ := g.CanShift(dir); can {
			return StateUndetermined
		}
	}
	return StateLost
}

func (g *Game) GetTile(x, y int) TileValue { return g.grid[y][x] }

func (g *Game) TileCount() int {
	out := 0
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			if g.grid[y][x] != 0 {
				out++
			}
		}
	}
	return out
}

// MarshalJSON defines how the Game struct is converted to JSON.
func (g Game) MarshalJSON() ([]byte, error) {
	// We wrap the private grid in an anonymous struct with an exported field
	return json.Marshal(struct {
		Grid [GridHeight][GridWidth]TileValue `json:"grid"`
	}{
		Grid: g.grid,
	})
}

// UnmarshalJSON defines how JSON data is loaded into the Game struct.
func (g *Game) UnmarshalJSON(data []byte) error {
	// Define a temporary shadow struct to catch the data
	temp := struct {
		Grid [GridHeight][GridWidth]TileValue `json:"grid"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Assign the captured data to our private field
	g.grid = temp.Grid
	return nil
}
