# BfH Server

_The Battle for Hermannia_ is a board game created as a gift by the father of [hermannm](https://github.com/hermannm), a developer of this project. This digital edition of _Hermannia_ is an online multiplayer implementation of the game. It consists of this server, written in Go, as well as a Unity client at [immerse-ntnu/bfh-client](https://github.com/immerse-ntnu/bfh-client).

## Package Structure

- The `game` package contains the game logic for _The Battle for Hermannia_, mainly the resolving of orders.
- The `boards` package contains the JSON files for the game's boards, and functions for deserializing them.
- The `lobby` package defines endpoints for finding, creating and joining lobbies, and manages WebSocket connections between players and lobbies. It is agnostic to the type of game played.
- The `messages` package defines the game-specific types of messages sent between client and server, as well as the logic for sorting incoming messages.
- The `app` package contains subpackages for each of the server's executables, and the common setup code for them.
  - The `main` package under `local` sets up a game server with a single, server-created lobby.
  - The `main` package under `public` sets up a game server where anyone can create their own lobbies through an open endpoint.
