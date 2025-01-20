package main

import (
    "github.com/notnil/chess"
    "fmt"
    "net/http"
    "os/exec"
    "io"
    "encoding/json"
    "strings"
    "time"
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
        err := fmt.Errorf("api request error: %w", err);
        panic(err)
    }
    defer r.Body.Close()
    body, err := io.ReadAll(r.Body)
    if err != nil {
        err := fmt.Errorf("reading json error: %w", err);
        panic(err)
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
        err := fmt.Errorf("pgn processing error: %w", err);
        panic(err)
    }

    return pgn
}

func BadMoves(last_game_pgn func(*chess.Game)) {
    // load stockfish
    stockfish := exec.Command("stockfish")
    stdin, err := stockfish.StdinPipe()
    if err != nil {
        err := fmt.Errorf("error loading stockfish stdin: %w", err)
        panic(err)
    }
    stdout, err := stockfish.StdoutPipe()
    if err != nil {
        err := fmt.Errorf("error loading stockfish stdout: %w", err)
        panic(err)
    }
    if err := stockfish.Start(); err != nil {
        err := fmt.Errorf("error starting stockfish: %w", err)
        panic(err)
    }
    defer stockfish.Wait()
    defer stockfish.Process.Kill()

    // functions to write and read from stockfish
	sendCommand := func(cmd string) {
		if _, err := stdin.Write([]byte(cmd + "\n")); err != nil {
            err := fmt.Errorf("Failed to send command to Stockfish: %w", err)
            panic(err)
		}
	}

	readOutput := func() string {
		buf := make([]byte, 2048)
		n, err := stdout.Read(buf)
		if err != nil {
            err := fmt.Errorf("Failed to read Stockfish output: %w", err)
            panic(err)
		}
		return string(buf[:n])
	}

    // load the pgn into a game and get moves
    game := chess.NewGame(last_game_pgn)
    moves := game.Moves()

    // Create new game
    walkthrough := chess.NewGame()
    // walkthrough new game and print each move
    fmt.Println(walkthrough.Position().Board().Draw())
    for i, move := range moves {
        err := walkthrough.Move(move)
        if err != nil {
            err := fmt.Errorf("walkthrough move boofed: %w", err)
            panic(err)
        }
        fmt.Println(walkthrough.Position().Board().Draw())
        if i == 10 {
            sendCommand("position fen " + walkthrough.FEN())
            sendCommand("go depth 10")
            time.Sleep(time.Millisecond * 200)
            fmt.Print(readOutput())
            break
        }
    }
}
