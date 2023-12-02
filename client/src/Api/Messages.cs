using System.Collections.Generic;
using CasusBelli.Client.Game;
using Godot;

namespace CasusBelli.Client.Api;

/// <summary>
/// Messages sent between the game client and server look like this:
/// <code>
/// {
///     "Tag": 4,
///     "Data": {"Faction": "green"}
/// }
/// </code>
/// ...where the "tag" field is one of the enum values defined in <see cref="MessageTag"/>, and
/// "data" is one of the message structs in <see cref="CasusBelli.Client.Api.Messages"/>.
/// </summary>
public record struct Message
{
    public required MessageTag Tag { get; set; }
    public required GodotObject Data { get; set; }
}

/// <summary>
/// Marker interface for message types that the client can receive from the server.
/// </summary>
public interface IReceivableMessage { }

/// <summary>
/// Marker interface for message types that the client can send to the server.
/// </summary>
public interface ISendableMessage { }

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
    public required string? Faction { get; set; }
}

/// <summary>
/// Message sent from a player when the lobby wants to start the game.
/// Requires that all players have selected a faction.
/// </summary>
public partial class StartGameMessage : GodotObject, ISendableMessage { }

/// <summary>
/// Message sent from server when the game starts.
/// </summary>
public partial class GameStartedMessage : GodotObject, IReceivableMessage
{
    public required Dictionary<string, Region> Board { get; set; }
}

/// <summary>
/// Message sent from server to client to signal that client should submit orders.
/// </summary>
public partial class OrderRequestMessage : GodotObject, IReceivableMessage
{
    public required Season Season { get; set; }
}

/// <summary>
/// Message sent from server to all clients when valid orders are received from a player.
/// Used to show who the server is waiting for.
/// </summary>
public partial class OrdersConfirmationMessage : GodotObject, IReceivableMessage
{
    public required string FactionThatSubmittedOrders { get; set; }
}

/// <summary>
/// Message sent from server to all clients when valid orders are received from all players.
/// </summary>
public partial class OrdersReceivedMessage : GodotObject, IReceivableMessage
{
    public required Dictionary<string, List<Order>> OrdersByFaction { get; set; }
}

/// <summary>
/// Message sent from server to all clients when a battle has begun.
/// resolving.
/// </summary>
public partial class BattleAnnouncementMessage : GodotObject, IReceivableMessage
{
    public required Battle Battle { get; set; }
}

/// <summary>
/// Message sent from server to all clients when a battle or danger zone crossing has finished
/// resolving.
/// </summary>
public partial class BattleResultsMessage : GodotObject, IReceivableMessage
{
    public required Battle Battle { get; set; }
}

/// <summary>
/// Message sent from server to all clients when the game is won.
/// </summary>
public partial class WinnerMessage : GodotObject, IReceivableMessage
{
    public required string WinningFaction { get; set; }
}

/// <summary>
/// Message sent from client when submitting orders.
/// </summary>
public partial class SubmitOrdersMessage : GodotObject, ISendableMessage
{
    public required List<Order> Orders { get; set; }
}

/// <summary>
/// Message sent from client when declaring who to support with their support order.
/// Forwarded by server to all clients to show who were given support.
/// </summary>
public partial class GiveSupportMessage : GodotObject, IReceivableMessage, ISendableMessage
{
    public required string EmbattledRegion { get; set; }

    /// <summary>
    /// Null if none were supported.
    /// </summary>
    public string? SupportedFaction { get; set; }
}

/// <summary>
/// Message sent from client when they roll the dice in a battle.
/// </summary>
public partial class DiceRollMessage : GodotObject, ISendableMessage { }
