# chess-game-review
review chess.com games with stockfish

currently grabs my most recent game on chess.com and prints it with a sssssimple bubbletea ui
## stockfish
stockfish now looks at the best moves at my position and tells me if I did not make the best move
## BROKEN:
it will by default print the white pieces as black and vice versa because the fon't is meant to be a black font. I fixed it by just going into the chess package source code and swapping the characters, but I need to find a permanent fix
