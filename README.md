# chess-game-review
review chess.com games with stockfish

currently grabs my most recent game on chess.com and prints it with a sssssimple bubbletea ui
## stockfish
stockfish now looks at the best moves at my position and tells me if I did not make the best move
## todo
- pretty up the game in the analysis tab
    - our name, opponents name
        - make persistent
    - +/- evaluation number (eval command in stockfish)
        - do this when we load the game for each position
- list of games from this month in the game selector
- be able to add new moves and eval them
## BROKEN:
it will by default print the white pieces as black and vice versa because the fon't is meant to be a black font. I fixed it by just going into the chess package source code and swapping the characters, but I need to find a permanent fix
## resources
- [chess.com published data api](https://www.chess.com/news/view/published-data-api#pubapi-endpoint-games)
- [stockfish commands](https://official-stockfish.github.io/docs/stockfish-wiki/UCI-&-Commands.html)
- [bubbletea examples](https://github.com/charmbracelet/bubbletea/tree/main/examples)
