using System.Collections.Generic;
using Godot;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Message sent from server when an error occurs.
/// </summary>
public partial class ErrorMessage : GodotObject, IReceivableMessage
{
    public required string Error { get; set; }
}

/// <summary>
/// Message sent from server to all clients when a player's status changes.
/// </summary>
public partial class PlayerStatusMessage : GodotObject, IReceivableMessage
{
    public required string Username { get; set; }
    public string? SelectedFaction { get; set; }
    public required bool ReadyToStartGame { get; set; }
}

/// <summary>
/// Message sent to a player when they join a lobby, to inform them about the game and other players.
/// </summary>
public partial class LobbyJoinedMessage : GodotObject, IReceivableMessage
{
    public required List<string> SelectableFactions { get; set; }
    public required List<PlayerStatusMessage> PlayerStatuses { get; set; }
}

/// <summary>
/// Message sent from client to select a faction to play for the game.
/// </summary>
public partial class SelectFactionMessage : GodotObject, ISendableMessage
{
    public required string Faction { get; set; }
}

/// <summary>
/// Message sent from a player when the lobby wants to start the game.
/// Requires that all players are ready.
/// </summary>
public partial class StartGameMessage : GodotObject, ISendableMessage { }
