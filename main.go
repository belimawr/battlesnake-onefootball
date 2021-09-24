package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	zlog "github.com/rs/zerolog/log"
)

type GameState struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

type Game struct {
	ID      string  `json:"id"`
	Ruleset Ruleset `json:"ruleset"`
	Timeout int32   `json:"timeout"`
}

type Ruleset struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Board struct {
	Height int           `json:"height"`
	Width  int           `json:"width"`
	Food   []Coord       `json:"food"`
	Snakes []Battlesnake `json:"snakes"`

	// Used in non-standard game modes
	Hazards []Coord `json:"hazards"`
}

type Battlesnake struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Health  int32   `json:"health"`
	Body    []Coord `json:"body"`
	Head    Coord   `json:"head"`
	Length  int32   `json:"length"`
	Latency string  `json:"latency"`

	// Used in non-standard game modes
	Shout string `json:"shout"`
	Squad string `json:"squad"`
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Response Structs

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type BattlesnakeMoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}

// HTTP Handlers

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := info()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		zlog.Printf("ERROR: Failed to encode info response, %s", err)
	}
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		zlog.Printf("ERROR: Failed to decode start json, %s", err)
		return
	}

	start(r.Context(), state)

	// Nothing to respond with here
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		zlog.Printf("ERROR: Failed to decode move json, %s", err)
		return
	}

	zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
		c = c.Str("GameID", state.Game.ID)
		c = c.Str("SnakeID", state.You.ID)
		c = c.Int("Turn", state.Turn)
		c = c.Stringer("url", r.URL)

		return c
	})

	response := move(r.Context(), state)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		zlog.Printf("ERROR: Failed to encode move response, %s", err)
		return
	}
}

func HandleEnd(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		zlog.Printf("ERROR: Failed to decode end json, %s", err)
		return
	}

	end(r.Context(), state)

	// Nothing to respond with here
}

// Main Entrypoint

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	logger := zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Replace the global logger
	log.SetFlags(0)
	log.SetOutput(logger)

	r := chi.NewRouter()
	r.Use(hlog.NewHandler(logger))
	r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	r.Get("/", HandleIndex)
	r.Post("/start", HandleStart)
	r.Post("/move", HandleMove)
	r.Post("/end", HandleEnd)

	logger.Info().Msgf("Starting Battlesnake Server at http://0.0.0.0:%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Fatal().Err(err).Msg("ListenAndServe error")
	}
}
