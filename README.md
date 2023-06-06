# BfH Game Server

_The Battle for Hermannia_ is a board game created as a gift by the father of [hermannm](https://github.com/hermannm), a developer of this project. This digital edition of _Hermannia_ is an online multiplayer implementation of the game. It consists of this server, written in Go, as well as a Unity client at [`immerse-ntnu/bfh-client`](https://github.com/immerse-ntnu/bfh-client).

## Project Structure

- `lobby` manages lobbies of players with persistent connections to the server, and messages between them
- `api` defines API endpoints for finding, creating and joining game lobbies
- `game` contains the main game loop
  - `gametypes` provides core data types used by other game packages, and operations on them
  - `ordervalidation` provides functions for gathering and validating game orders from players
  - `orderresolving` contains the bulk of the game logic: resolving player orders on the board
  - `boardconfig` contains JSON files for board setups that can be played, and functions for parsing them
- Package `main` at the project root contains a CLI application for launching the game server
- `magefiles` contains build scripts using [Mage](https://magefile.org/)

## Credits

- Tomas H.V. Mørkrid for creating the original board game
- [Lars-over](https://github.com/Lars-over), [kristley](https://github.com/kristley), [bjorvik](https://github.com/bjorvik) and [ragnsol](https://github.com/ragnsol) of [Immerse NTNU](https://github.com/immerse-ntnu) for their work on the [game client](https://github.com/immerse-ntnu/bfh-client)
- [gorilla/websocket](https://github.com/gorilla/websocket) for the Go WebSocket package
  - _Copyright (c) 2013 The Gorilla WebSocket Authors, [BSD 2-clause license](https://github.com/gorilla/websocket/blob/master/LICENSE)_
