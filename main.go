package main

import (
    "github.com/notnil/chess"
    "fmt"
    "net/http"
    "os"
    // "os/exec"
    "io"
    "encoding/json"
    "strings"
)

func main() {
    // get pgn of my most recent game
    last_game_pgn := LastGamePgn()

    // analyze each move with stockfish and point out my bad moves
    BadMoves(last_game_pgn)
}

func LastGamePgn() func(*chess.Game) {
    // sample http request with my games from this month
    // hardcoded this month for now
    r, err := http.Get("https://api.chess.com/pub/player/ggumption/games/2025/01")
    if err != nil {
        fmt.Errorf("api request error: %w", err);
    }
    defer r.Body.Close()
    body, err := io.ReadAll(r.Body)
    if err != nil {
        fmt.Errorf("reading json error: %w", err);
    }

    // process the json
    var d interface{}
    json.Unmarshal(body, &d)
    m := d.(map[string]interface{}) // processes a string json field

    // list of game PGNs
    games := m["games"].([]interface{}) // processes a list

    // recent game:
    recent_game := games[len(games) - 1].(map[string]interface{}) // string map for fields
    gr := strings.NewReader(recent_game["pgn"].(string)) // convert pgn field to string reader for PGN function

    // process the pgn
    pgn, err := chess.PGN(gr)
    if err != nil {
        fmt.Errorf("pgn processing error: %w", err);
    }

    return pgn
}

func BadMoves(last_game_pgn func(*chess.Game)) {
    // load the pgn into a game and get moves
    game := chess.NewGame(last_game_pgn)
    moves := game.Moves()

    // Create new game from the reversed position
    walkthrough := chess.NewGame()
    // walkthrough new game and print each move
    fmt.Println(walkthrough.Position().Board().Draw())
    for i, move := range moves {
        err := walkthrough.Move(move)
        if err != nil {
            fmt.Println("move in walkthrough boofed")
            os.Exit(1)
        }
        fmt.Println(walkthrough.Position().Board().Draw())
        if i == 10 {
            fmt.Print(walkthrough.FEN())
            break
        }
    }
    
}
