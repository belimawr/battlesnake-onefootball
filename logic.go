package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"context"
	"math/rand"

	"github.com/rs/zerolog"
)

// This function is called when you register your Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your Battlesnake URL in browser you should see this data.
func info() BattlesnakeInfoResponse {
	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "Tiago Queiroz", // TODO: Your Battlesnake username
		Color:      "#babaca",       // TODO: Personalize
		Head:       "snowman",       // TODO: Personalize
		Tail:       "coffee",        // TODO: Personalize
	}
}

// This function is called everytime your Battlesnake is entered into a game.
// The provided GameState contains information about the game that's about to be played.
// It's purely for informational purposes, you don't have to make any decisions here.
func start(ctx context.Context, state GameState) {
	zerolog.Ctx(ctx).Info().Msgf("%s START", state.Game.ID)
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(ctx context.Context, state GameState) {
	zerolog.Ctx(ctx).Info().Msgf("%s END", state.Game.ID)
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(ctx context.Context, state GameState) BattlesnakeMoveResponse {
	logger := zerolog.Ctx(ctx)
	me := state.You
	possibleMoves := map[string]bool{
		"up":    true,
		"down":  true,
		"left":  true,
		"right": true,
	}

	// Step 0: Don't let your Battlesnake move back in on it's own neck
	myHead := state.You.Body[0] // Coordinates of your head
	myNeck := state.You.Body[1] // Coordinates of body piece directly behind your head (your "neck")
	if myNeck.X < myHead.X {
		possibleMoves["left"] = false
	} else if myNeck.X > myHead.X {
		possibleMoves["right"] = false
	} else if myNeck.Y < myHead.Y {
		possibleMoves["down"] = false
	} else if myNeck.Y > myHead.Y {
		possibleMoves["up"] = false
	}

	// TODO: Step 1 - Don't hit walls.
	// Use information in GameState to prevent your Battlesnake from moving beyond the boundaries of the board.
	boardWidth := state.Board.Width
	boardHeight := state.Board.Height

	if myHead.X == boardWidth-1 {
		possibleMoves["right"] = false
	}

	if myHead.X == 0 {
		possibleMoves["left"] = false
	}

	if myHead.Y == boardHeight-1 {
		possibleMoves["up"] = false
	}

	if myHead.Y == 0 {
		possibleMoves["down"] = false
	}

	// Places to avoid
	snakes := map[Coord]struct{}{}
	// Add myself
	for _, c := range me.Body {
		snakes[c] = struct{}{}
	}

	// Add all other snakes and allow for head-to-head
	// if we can win
	for _, s := range state.Board.Snakes {
		if me.ID == s.ID {
			continue
		}

		// I win on head to head collisions
		// Add the other snake's body, but not its head
		if me.Health > s.Health {
			for i := 1; i < len(s.Body); i++ {
				snakes[s.Body[i]] = struct{}{}
			}
			continue
		}

		for _, p := range s.Body {
			snakes[p] = struct{}{}
		}
	}

	// Add possible next moves fom all snakes
	// but ourselves
	for _, s := range state.Board.Snakes {
		if me.ID == s.ID {
			continue
		}

		for _, p := range nextPossibleMoves(s.Head) {
			snakes[p] = struct{}{}
		}
	}

	// For each possible move, select the safe ones
	for p := range snakes {
		for _, m := range []string{"up", "down", "left", "right"} {
			switch m {
			case "up":
				nextHead := myHead
				nextHead.Y++
				if p == nextHead {
					possibleMoves["up"] = false

				}
				break
			case "down":
				nextHead := myHead
				nextHead.Y--
				if p == nextHead {
					possibleMoves["down"] = false

				}
				break
			case "left":
				nextHead := myHead
				nextHead.X--
				if p == nextHead {
					possibleMoves["left"] = false
				}
				break
			case "right":
				nextHead := myHead
				nextHead.X++
				if p == nextHead {
					possibleMoves["right"] = false
				}
				break
			}
		}
	}

	// TODO: Step 3 - Don't collide with others.
	// Use information in GameState to prevent your Battlesnake from colliding with others.

	// TODO: Step 4 - Find food.
	// Use information in GameState to seek out and find food.

	// Finally, choose a move from the available safe moves.
	// TODO: Step 5 - Select a move to make based on strategy, rather than random.

	nextMove := findNextMove(ctx, state, possibleMoves)
	logger.Info().Msgf("MOVE: %s", nextMove)

	return BattlesnakeMoveResponse{
		Move: nextMove,
	}
}

func findNextMove(ctx context.Context, state GameState, possibleMoves map[string]bool) string {
	me := state.You
	// Find food
	if me.Health < 20 {
		zerolog.Ctx(ctx).Debug().Msg("find food!")
		return gotoFood(ctx, state, possibleMoves)
	}

	// loop up and down
	if possibleMoves["up"] {
		return "up"
	}

	if possibleMoves["left"] {
		return "left"
	}

	if possibleMoves["down"] {
		return "down"
	}

	if possibleMoves["right"] {
		return "right"
	}

	return "down"
}

func gotoFood(ctx context.Context, state GameState, possibleMoves map[string]bool) string {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msgf("%s TURN %d: Finding food", state.Game.ID, state.Turn)

	myHead := state.You.Head
	// If there is no food, move randomly
	if len(state.Board.Food) == 0 {
		return randomMove(ctx, state, possibleMoves)
	}

	for _, food := range state.Board.Food {
		if myHead.X > food.X {
			if possibleMoves["left"] {
				return "left"
			}
			continue
		}
		if myHead.X < food.X {
			if possibleMoves["right"] {
				return "right"
			}
			continue
		}

		if myHead.Y > food.Y {
			if possibleMoves["down"] {
				return "down"
			}
			continue
		}
		if myHead.Y < food.Y {
			if possibleMoves["up"] {
				return "up"
			}
			continue
		}
	}

	return randomMove(ctx, state, possibleMoves)
}

func randomMove(ctx context.Context, state GameState, possibleMoves map[string]bool) string {
	safeMoves := []string{}

	for move, isSafe := range possibleMoves {
		if isSafe {
			safeMoves = append(safeMoves, move)
		}
	}

	if len(safeMoves) == 0 {
		zerolog.Ctx(ctx).Info().Msg("NO SAFE MOVES! Going down")
		return "down"
	}

	zerolog.Ctx(ctx).Info().Msg("Random move!")
	return safeMoves[rand.Intn(len(safeMoves))]
}

func nextPossibleMoves(head Coord) []Coord {
	moves := []Coord{}
	moves = append(moves, Coord{X: head.X + 1, Y: head.Y})
	moves = append(moves, Coord{X: head.X - 1, Y: head.Y})
	moves = append(moves, Coord{X: head.X, Y: head.Y + 1})
	moves = append(moves, Coord{X: head.X, Y: head.Y - 1})

	return moves
}

//Nice game: https://play.battlesnake.com/g/c22c39c1-c13c-4772-844a-c7b5816c3460/?turn=308
