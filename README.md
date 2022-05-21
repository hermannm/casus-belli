# BfH Server

_The Battle for Hermannia_ is a board game created as a gift by the father of [hermannm](https://github.com/hermannm), a developer of this project. This digital edition of _Hermannia_ is an online multiplayer implementation of the game. It consists of this server, written in Go, as well as a Unity client at [immerse-ntnu/bfh-client](https://github.com/immerse-ntnu/bfh-client).

## Package Structure

- Package `lobby` defines API endpoints for finding, creating and joining game lobbies, and manages WebSocket connections between players and the server. It is agnostic to the type of game played.
- Package `game` implements _The Battle for Hermannia_ as a game for `lobby`. It contains all subpackages specific to this game.
  - Package `board` contains the main game logic, mainly the resolving of orders on the board.
  - Package `boardsetup` contains the JSON files for the game's boards, and functions for deserializing them.
  - Package `messages` defines the game-specific types of messages sent between client and server, as well as the logic for sorting incoming messages.
  - Package `validation` contains functions for validating player input.
- `cmd` contains packages for each of the server's executables.
  - Package `main` under `local` sets up a game server with a single, server-created lobby.
  - Package `main` under `public` sets up a game server where anyone can create their own lobbies through an open endpoint.
- Package `server` at the project root contains the common setup code for the server's executables, such as the configured games.

## Credits

- Tomas H.V. Mørkrid for creating the original board game
- [Lars-over](https://github.com/Lars-over), [kristley](https://github.com/kristley) and [bjorvik](https://github.com/bjorvik) of [Immerse NTNU](https://github.com/immerse-ntnu) for their support, as well as their work on [the client](https://github.com/immerse-ntnu/bfh-client)
- [gorilla/websocket](https://github.com/gorilla/websocket) for the Go WebSocket package
  - _Copyright (c) 2013 The Gorilla WebSocket Authors, [BSD 2-clause license](https://github.com/gorilla/websocket/blob/master/LICENSE)_
