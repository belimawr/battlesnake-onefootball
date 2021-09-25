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
func info(ctx context.Context) BattlesnakeInfoResponse {
	zerolog.Ctx(ctx).Info().Msg("Sending snake info")
	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "Tiago Queiroz",
		Color:      "#1a0a74",
		Head:       "snowman",
		Tail:       "coffee",
	}
}

// This function is called everytime your Battlesnake is entered into a game.
// The provided GameState contains information about the game that's about to be played.
// It's purely for informational purposes, you don't have to make any decisions here.
func start(ctx context.Context, state GameState) {
	zerolog.Ctx(ctx).Info().Msg("Starting game!")
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(ctx context.Context, state GameState) {
	zerolog.Ctx(ctx).Info().Msg("The game has ended")
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(ctx context.Context, state GameState) BattlesnakeMoveResponse {
	logger := zerolog.Ctx(ctx)

	possibleMoves := safeMoves(state, true, true)
	safeMoves := 0
	for _, isSafe := range possibleMoves {
		if isSafe {
			safeMoves++
		}
	}

	if safeMoves == 0 {
		logger.Debug().Msg("no safe moves, trying a less restrictive approach") //
		nextMove := randomEmptySquare(ctx, state)
		logger.Info().Msgf("MOVE: %s", nextMove)

		return BattlesnakeMoveResponse{
			Move: nextMove,
		}
	}

	nextMove := findNextMove(ctx, state, possibleMoves)
	logger.Info().Msgf("MOVE: %s", nextMove)

	return BattlesnakeMoveResponse{
		Move: nextMove,
	}
}

func randomEmptySquare(ctx context.Context, state GameState) string {
	zerolog.Ctx(ctx).Debug().Msg("Random empty square")

	possibleMoves := safeMoves(state, true, false)

	return randomMove(ctx, state, possibleMoves)
}

func findNextMove(ctx context.Context, state GameState, possibleMoves map[string]bool) string {
	me := state.You
	// Find food
	if me.Health < 75 {
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

// safeMoves returns all safe moves.
// - allowHeadColision: Allows for head collisions when we're bigger
// - predictMoves: makes unsae any move that could hit any
//   snake's head on the next move
func safeMoves(
	state GameState,
	allowHeadColision, predictMoves bool,
) map[string]bool {

	possibleMoves := map[string]bool{
		"up":    true,
		"down":  true,
		"left":  true,
		"right": true,
	}

	me := state.You
	myHead := me.Head
	boardWidth := state.Board.Width
	boardHeight := state.Board.Height

	// Avoid Walls
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

	// Avoid hitting other noGoCoords
	noGoCoords := map[Coord]struct{}{}
	// Add ourselves
	for _, c := range me.Body {
		noGoCoords[c] = struct{}{}
	}

	// Add all other snakes and allow for head-to-head
	// if we can win
	for _, s := range state.Board.Snakes {
		if me.ID == s.ID {
			continue
		}

		// Add the whole snake's body
		for _, p := range s.Body {
			noGoCoords[p] = struct{}{}
		}

		if allowHeadColision {
			// If we win on head to head collisions
			// remove the other snake's head from noGoCoords
			if me.Health > s.Health {
				delete(noGoCoords, s.Head)
			}
		}
	}

	if predictMoves {
		// Add possible next moves fom all snakes
		// but ourselves
		for _, s := range state.Board.Snakes {
			if me.ID == s.ID {
				continue
			}

			for _, p := range nextPossibleMoves(s.Head) {
				noGoCoords[p] = struct{}{}
			}
		}
	}

	// For each possible move, verify which ones are safe
	for p := range noGoCoords {
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

	return possibleMoves
}

//Nice game: https://play.battlesnake.com/g/c22c39c1-c13c-4772-844a-c7b5816c3460/?turn=308
//length 30: https://play.battlesnake.com/g/60269951-977a-47c9-b6bd-a5e5e002030c/?turn=243
