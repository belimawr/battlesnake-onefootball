package main

// This file can be a nice home for your Battlesnake logic and related helper functions.
//
// We have started this for you, with a function to help remove the 'neck' direction
// from the list of possible moves!

import (
	"log"
	"math/rand"
)

// This function is called when you register your Battlesnake on play.battlesnake.com
// See https://docs.battlesnake.com/guides/getting-started#step-4-register-your-battlesnake
// It controls your Battlesnake appearance and author permissions.
// For customization options, see https://docs.battlesnake.com/references/personalization
// TIP: If you open your Battlesnake URL in browser you should see this data.
func info() BattlesnakeInfoResponse {
	log.Println("INFO")
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
func start(state GameState) {
	log.Printf("%s START\n", state.Game.ID)
}

// This function is called when a game your Battlesnake was in has ended.
// It's purely for informational purposes, you don't have to make any decisions here.
func end(state GameState) {
	log.Printf("%s END\n\n", state.Game.ID)
}

// This function is called on every turn of a game. Use the provided GameState to decide
// where to move -- valid moves are "up", "down", "left", or "right".
// We've provided some code and comments to get you started.
func move(state GameState) BattlesnakeMoveResponse {
	// fmt.Println("Sankes")
	// for _, snake := range state.Board.Snakes {
	// 	fmt.Printf("ID: %s, Body: %#v\n", snake.ID, snake.Body)
	// }

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

	log.Print("Board   ", boardWidth, boardHeight)
	log.Print("My head ", state.You.Head.X, state.You.Head.Y)

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

	// TODO: Step 2 - Don't hit yourself.
	// Use information in GameState to prevent your Battlesnake from colliding with itself.
	// mybody := state.You.Body

	// For each possible move do something:
	for _, m := range []string{"up", "down", "left", "right"} {
		switch m {
		case "up":
			nextHead := myHead
			nextHead.Y++
			for _, s := range state.Board.Snakes {
				for _, p := range s.Body {
					if p == nextHead {
						possibleMoves["up"] = false
						break
					}
				}
			}
			break
		case "down":
			nextHead := myHead
			nextHead.Y--
			for _, s := range state.Board.Snakes {
				for _, p := range s.Body {
					if p == nextHead {
						possibleMoves["down"] = false
						break
					}
				}
			}
			break
		case "left":
			nextHead := myHead
			nextHead.X--
			for _, s := range state.Board.Snakes {
				for _, p := range s.Body {
					if p == nextHead {
						possibleMoves["left"] = false
						break
					}
				}
			}
			break
		case "right":
			nextHead := myHead
			nextHead.X++
			for _, s := range state.Board.Snakes {
				for _, p := range s.Body {
					if p == nextHead {
						possibleMoves["right"] = false
						break
					}
				}
			}
			break
		}
	}

	// TODO: Step 3 - Don't collide with others.
	// Use information in GameState to prevent your Battlesnake from colliding with others.

	// TODO: Step 4 - Find food.
	// Use information in GameState to seek out and find food.

	// Finally, choose a move from the available safe moves.
	// TODO: Step 5 - Select a move to make based on strategy, rather than random.

	nextMove := findNextMove(state, possibleMoves)
	log.Printf("%s TURN %d: %s\n", state.Game.ID, state.Turn, nextMove)

	return BattlesnakeMoveResponse{
		Move: nextMove,
	}
}

func findNextMove(state GameState, possibleMoves map[string]bool) string {
	me := state.You
	// Find food
	if me.Health < 30 {
		log.Println("find food!")
		return gotoFood(state, possibleMoves)
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

func gotoFood(state GameState, possibleMoves map[string]bool) string {
	log.Printf("%s TURN %d: Finding food\n", state.Game.ID, state.Turn)
	myHead := state.You.Head
	// If there is no food, move randomly
	if len(state.Board.Food) == 0 {
		return randomMove(state, possibleMoves)
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

	return randomMove(state, possibleMoves)
}

func randomMove(state GameState, possibleMoves map[string]bool) string {
	safeMoves := []string{}

	for move, isSafe := range possibleMoves {
		if isSafe {
			safeMoves = append(safeMoves, move)
		}
	}

	if len(safeMoves) == 0 {
		log.Printf("%s TURN %d: NO SAFE MOVES! Going down\n", state.Game.ID, state.Turn)
		return "down"
	}

	log.Printf("%s TURN %d: Random move!\n", state.Game.ID, state.Turn)
	return safeMoves[rand.Intn(len(safeMoves))]
}

//Nice game: https://play.battlesnake.com/g/c22c39c1-c13c-4772-844a-c7b5816c3460/?turn=308
