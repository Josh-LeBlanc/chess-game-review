package main

import (
    "github.com/notnil/chess"
    "fmt"
    "net/http"
    "os"
    "io"
    "encoding/json"
    "strings"
)

func main() {
    // sample http request with my games from this month
    r, err := http.Get("https://api.chess.com/pub/player/ggumption/games/2025/01")
    if err != nil {
        fmt.Println("api request bunked")
        os.Exit(1)
    }
    defer r.Body.Close()
    body, err := io.ReadAll(r.Body)

    // process the json
    var d interface{}
    json.Unmarshal(body, &d)
    m := d.(map[string]interface{}) // processes s string json field

    // list of game PGNs
    games := m["games"].([]interface{}) // processes a list

    // first game:
    game1 := games[0].(map[string]interface{}) // string map for fields
    gr := strings.NewReader(game1["pgn"].(string)) // convert pgn field to string reader for PGN function

    fmt.Print(game1["pgn"])

    // process the pgn
    pgn, err := chess.PGN(gr)
    if err != nil {
        fmt.Println("pgn error")
        os.Exit(1)
    }

    // load the pgn into a game and get moves
    game := chess.NewGame(pgn)
    moves := game.Moves()

    // // reverse positions because my white and black pieces are the wrong colors :(
    // reversedFEN := "RNBQKBNR/PPPPPPPP/8/8/8/8/pppppppp/rnbqkbnr w KQkq - 0 1"
    // fen, err := chess.FEN(reversedFEN)
    // if err != nil {
    //     fmt.Println("fen boofed")
    //     os.Exit(1)
    // }
    // walkthrough := chess.NewGame(fen)

    // Create new game from the reversed position
    walkthrough := chess.NewGame()
    fmt.Println(walkthrough.Position().Board().Draw())
    for i, move := range moves {
        err := walkthrough.Move(move)
        if err != nil {
            fmt.Println("move in walkthrough boofed")
            os.Exit(1)
        }
        fmt.Println(walkthrough.Position().Board().Draw())
        if i == 5 { break }
    }
}

