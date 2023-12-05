# Casus Belli

_Casus Belli_ ("justification for war") is a game of strategy, diplomacy and battle. It was
originally created as a gift by the father of [hermannm](https://github.com/hermannm), a developer
of this project. This digital edition of the board game consists of a server written in Go, and a
client built with the Godot game engine and C#.

## Project Structure

- `server`
  - `lobby` manages lobbies of players with WebSocket connections to the server, and messages
    between them
  - `api` defines API endpoints for finding, creating and joining game lobbies
  - `game` implements the game logic (resolving player orders on the board)
- `client`
  - `src` contains C# game scripts
  - `scenes` contains Godot scene files
  - `assets` contains textures, images, icons and fonts

## Development Setup

### Server

- Install Go: https://go.dev/
- Run the server: navigate to `casus-belli/server`, then `go run .`
  - To run in single-lobby mode for local server hosting: `go run . -local`
- To run cross-compilation build script, install Mage: https://magefile.org/
  - Run `mage crosscompile` (in `casus-belli/server`) to compile server for all supported OSes

### Client

- Install Godot 4.1 or newer, with .NET support: https://godotengine.org/
- Install .NET 7.0: https://dotnet.microsoft.com/en-us/download
- Install .NET project tools: navigate to `casus-belli/client`, and run `dotnet tool restore`
- (Recommended) Install the Csharpier formatter plugin for your IDE, and enable format-on-save:
  https://csharpier.com/docs/Editors

## Credits

- Tomas H.V. Mørkrid for creating the original board game
- [kristley](https://github.com/kristley), [bjorvik](https://github.com/bjorvik),
  [Lars-over](https://github.com/Lars-over) and [ragnsol](https://github.com/ragnsol) of
  [Immerse NTNU](https://github.com/immerse-ntnu) for their contributions to the game client
- [gorilla/websocket](https://github.com/gorilla/websocket) for the Go WebSocket package
