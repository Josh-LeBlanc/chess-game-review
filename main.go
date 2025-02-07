package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Josh-LeBlanc/chess-game-review/display"
	"github.com/notnil/chess"
)

func main() {
	// get pgn of my most recent game
	last_game_pgn, md, _ := LastGamePgn()

	// analyze each move with stockfish and point out my bad moves
	// BadMoves(last_game_pgn, white)

	game := chess.NewGame(last_game_pgn)

	display.Display(game, md)

	// this has saved our month of games in a text file
	// SaveMyRecentApiReq()
}

func LastGamePgn() (func(*chess.Game), display.GameMetadata, bool) {
	body := ReadMyRecentApiReq()

	// process the json
	var d interface{}
	json.Unmarshal(body, &d)
	m := d.(map[string]interface{}) // processes a string json field

	// list of game PGNs
	games := m["games"].([]interface{}) // processes a list

	// recent game:
	recent_game := games[len(games)-1].(map[string]interface{}) // string map for fields
	gr := strings.NewReader(recent_game["pgn"].(string))        // convert pgn field to string reader for PGN function

	// process the pgn
	pgn, err := chess.PGN(gr)
	if err != nil {
		err := fmt.Errorf("pgn processing error: %w", err)
		panic(err)
	}

	// retrieve my color
	var white bool
	if string(strings.Split(recent_game["pgn"].(string), "White")[1][3:8]) == "Gumpt" {
		white = true
	} else {
		white = false
	}

	// get game metadata
	w := recent_game["white"].(map[string]interface{})
	b := recent_game["black"].(map[string]interface{})
	md := display.GameMetadata{
		White: w["username"].(string) + " (" + fmt.Sprintf("%.0f", w["rating"].(float64)) + ")",
		Black: b["username"].(string) + " (" + fmt.Sprintf("%.0f", b["rating"].(float64)) + ")",
	}

	return pgn, md, white
}

func MyRecentApiReq() []byte {
	// sample http request with my games from this month
	// hardcoded this month for now
	my_recent_month_url := "https://api.chess.com/pub/player/ggumption/games/" + time.Now().Format("2006/01")
	r, err := http.Get(my_recent_month_url)
	if err != nil {
		err := fmt.Errorf("api request error: %w", err)
		panic(err)
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err := fmt.Errorf("reading json error: %w", err)
		panic(err)
	}

	return body
}

func SaveMyRecentApiReq() {
	body := MyRecentApiReq()

	filename := "saved_api_requests/me-" + time.Now().Format("01-06") + ".txt"

	err := os.WriteFile(filename, body, 0644)
	if err != nil {
		err := fmt.Errorf("writing api data file error: %w", err)
		panic(err)
	}
}

func ReadMyRecentApiReq() []byte {
	filename := "saved_api_requests/me-" + time.Now().Format("01-06") + ".txt"

	body, err := os.ReadFile(filename)
	if err != nil {
		err := fmt.Errorf("reading api data file error: %w", err)
		panic(err)
	}
	return body
}

func BadMoves(last_game_pgn func(*chess.Game), white bool) {
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

	readStockfishOutput := func() string {
		buf := make([]byte, 3072)
		n, err := stdout.Read(buf)
		if err != nil {
			err := fmt.Errorf("Failed to read Stockfish output: %w", err)
			panic(err)
		}
		return string(buf[:n])
	}

	// set stockfish options
	sendCommand("setoption name Threads value 4")

	// load the pgn into a game and get moves
	game := chess.NewGame(last_game_pgn)
	moves := game.Moves()

	// Create new game
	walkthrough := chess.NewGame()

	// Configuration constants
	const (
		analysisDepth  = 15  // Increased from 10 for better accuracy
		moveLimit      = 20  // Analyze more moves
		evalThreshold  = 100 // Centipawn threshold for "bad" moves
		engineWaitTime = 500 * time.Millisecond
	)

	// Helper function to get evaluation from Stockfish output
	getEvaluation := func(output string) (int, error) {
		// Look for "score cp" in the output
		if idx := strings.Index(output, "score cp "); idx != -1 {
			scoreStr := strings.Fields(output[idx:])[2]
			score, err := strconv.Atoi(scoreStr)
			return score, err
		}
		return 0, fmt.Errorf("no evaluation found")
	}

	// walkthrough new game and analyze each move
	for i, move := range moves {
		if i >= moveLimit {
			break
		}

		err := walkthrough.Move(move)
		if err != nil {
			fmt.Printf("Warning: Error making move %d: %v\n", i, err)
			continue
		}

		// Only analyze player's moves
		if (i%2 == 0) != white {
			// Get position evaluation before the move
			sendCommand("position fen " + walkthrough.Position().String())
			sendCommand(fmt.Sprintf("go depth %d", analysisDepth))

			time.Sleep(engineWaitTime)
			beforeOut := readStockfishOutput()
			beforeEval, _ := getEvaluation(beforeOut)
			beforeBest := strings.Split(strings.Split(beforeOut, "bestmove")[1], " ")[1]

			// If move differs from engine suggestion, calculate evaluation difference
			if fmt.Sprintf("%s", moves[i]) != beforeBest {
				// Get evaluation after the move
				sendCommand("position fen " + walkthrough.Position().String())
				sendCommand(fmt.Sprintf("go depth %d", analysisDepth))

				time.Sleep(engineWaitTime)
				afterOut := readStockfishOutput()
				afterEval, _ := getEvaluation(afterOut)

				evalDiff := beforeEval - afterEval
				if evalDiff > evalThreshold {
					fmt.Printf("\nSignificant mistake found at move %d:\n", (i/2)+1)
					fmt.Printf("Position evaluation before: %d\n", beforeEval)
					fmt.Printf("Position evaluation after: %d\n", afterEval)
					fmt.Printf("Your move: %s\n", moves[i])
					fmt.Printf("Suggested move: %s\n", beforeBest)
					fmt.Printf("Evaluation difference: %d centipawns\n", evalDiff)
					fmt.Println(walkthrough.Position().Board().Draw())
					fmt.Println()
				}
			}
		}
	}
}
