# chess-game-review
review chess.com games with stockfish

alright right now we just fetch my first game in january and print the first few moves
## stockfish
download stockfish

__change of plans:__
looks like chess package has a better stockfish api than i thought, going to try to use that before making my own
## BROKEN:
it will by default print the white pieces as black and vice versa because the fon't is meant to be a black font. I fixed it by just going into the chess package source code and swapping the characters, but I need to find a permanent fix
