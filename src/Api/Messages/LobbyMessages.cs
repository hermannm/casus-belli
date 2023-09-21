using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Message sent from server when an error occurs.
/// </summary>
public struct ErrorMessage : IReceivableMessage
{
    /// <summary>
    /// The error message.
    /// </summary>
    [JsonPropertyName("error")]
    public required string Error;
}

/// <summary>
/// Message sent from server to all clients when a player's status changes.
/// </summary>
public struct PlayerStatusMessage : IReceivableMessage
{
    /// <summary>
    /// The user's chosen display name.
    /// </summary>
    [JsonPropertyName("username")]
    public required string Username;

    /// <summary>
    /// The user's selected game ID.
    /// Null if not selected yet.
    /// </summary>
    [JsonPropertyName("gameId")]
    public string? GameId;

    /// <summary>
    /// Whether the user is ready to start the game.
    /// </summary>
    [JsonPropertyName("ready")]
    public required bool Ready;
}

/// <summary>
/// Message sent to a player when they join a lobby, to inform them about other players.
/// </summary>
public struct LobbyJoinedMessage : IReceivableMessage
{
    /// <summary>
    /// IDs that the player may select from for this lobby's game.
    /// Returns all game IDs, though some may already be taken by other players in the lobby.
    /// </summary>
    [JsonPropertyName("gameIds")]
    public required List<string> GameIds;

    /// <summary>
    /// Info about each other player in the lobby.
    /// </summary>
    [JsonPropertyName("playerStatuses")]
    public required List<PlayerStatusMessage> PlayerStatuses;
}

/// <summary>
/// Message sent from client when they want to select a game ID.
/// </summary>
public struct SelectGameIdMessage : ISendableMessage
{
    /// <summary>
    /// The ID that the player wants to select for the game.
    /// Will be rejected if already selected by another player.
    /// </summary>
    [JsonPropertyName("gameId")]
    public required string GameId;
}

/// <summary>
/// Message sent from client to mark themselves as ready to start the game.
/// Requires game ID being selected.
/// </summary>
public struct ReadyMessage : ISendableMessage
{
    /// <summary>
    /// Whether the player is ready to start the game.
    /// </summary>
    [JsonPropertyName("ready")]
    public required bool Ready;
}

/// <summary>
/// Message sent from a player when the lobby wants to start the game.
/// Requires that all players are ready.
/// </summary>
public struct StartGameMessage : ISendableMessage { }
