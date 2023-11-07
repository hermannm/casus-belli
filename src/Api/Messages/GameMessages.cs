using System.Collections.Generic;
using System.Text.Json.Serialization;
using Godot;
using Immerse.BfhClient.Api.GameTypes;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Message sent from server when asking a supporting player who to support in an embattled region.
/// </summary>
public partial class SupportRequestMessage : GodotObject, IReceivableMessage
{
    /// <summary>
    /// The region from which support is asked, where the asked player should have a support order.
    /// </summary>
    [JsonPropertyName("supportingRegion")]
    public required string SupportingRegion { get; set; }

    /// <summary>
    /// List of possible players to support in the battle.
    /// </summary>
    [JsonPropertyName("supportablePlayers")]
    public required List<string> SupportablePlayers { get; set; }
}

/// <summary>
/// Message sent from server to client to signal that client should submit orders.
/// </summary>
public partial class OrderRequestMessage : GodotObject, IReceivableMessage { }

/// <summary>
/// Message sent from server to all clients when valid orders are received from all players.
/// </summary>
public partial class OrdersReceivedMessage : GodotObject, IReceivableMessage
{
    /// <summary>
    /// Maps a player's ID to their submitted orders.
    /// </summary>
    [JsonPropertyName("playerOrders")]
    public required Dictionary<string, List<Order>> PlayerOrders { get; set; }
}

/// <summary>
/// Message sent from server to all clients when valid orders are received from a player.
/// Used to show who the server is waiting for.
/// </summary>
public partial class OrdersConfirmationMessage : GodotObject, IReceivableMessage
{
    /// <summary>
    /// The player who submitted orders.
    /// </summary>
    [JsonPropertyName("player")]
    public required string Player { get; set; }
}

/// <summary>
/// Message sent from server to all clients when a battle result is calculated.
/// </summary>
public partial class BattleResultsMessage : GodotObject, IReceivableMessage
{
    /// <summary>
    /// The relevant battle result.
    /// </summary>
    [JsonPropertyName("battles")]
    public required List<Battle> Battles { get; set; }
}

/// <summary>
/// Message sent from server to all clients when the game is won.
/// </summary>
public partial class WinnerMessage : GodotObject, IReceivableMessage
{
    /// <summary>
    /// Player tag of the game's winner.
    /// </summary>
    [JsonPropertyName("winner")]
    public required string Winner { get; set; }
}

/// <summary>
/// Message sent from client when submitting orders.
/// </summary>
public partial class SubmitOrdersMessage : GodotObject, ISendableMessage
{
    /// <summary>
    /// List of submitted orders.
    /// </summary>
    [JsonPropertyName("orders")]
    public required List<Order> Orders { get; set; }
}

/// <summary>
/// Message sent from client when declaring who to support with their support order.
/// Forwarded by server to all clients to show who were given support.
/// </summary>
public partial class GiveSupportMessage : GodotObject, IReceivableMessage, ISendableMessage
{
    /// <summary>
    /// Name of the region in which the support order is placed.
    /// </summary>
    [JsonPropertyName("supportingRegion")]
    public required string SupportingRegion { get; set; }

    /// <summary>
    /// ID of the player in the destination region to support.
    /// Null if none were supported.
    /// </summary>
    [JsonPropertyName("supportedPlayer")]
    public string? SupportedPlayer { get; set; }
}

/// <summary>
/// Message passed from the client during winter council voting.
/// Used for the throne expansion.
/// </summary>
public partial class WinterVoteMessage : GodotObject, ISendableMessage
{
    /// <summary>
    /// ID of the player that the submitting player votes for.
    /// </summary>
    [JsonPropertyName("player")]
    public required string Player { get; set; }
}

/// <summary>
/// Message passed from the client with the swordMsg to declare where they want to use it.
/// Used for the throne expansion.
/// </summary>
public partial class SwordMessage : GodotObject, ISendableMessage
{
    /// <summary>
    /// Name of the region in which the player wants to use the sword in battle.
    /// </summary>
    [JsonPropertyName("region")]
    public required string Region { get; set; }

    /// <summary>
    /// Index of the battle in which to use the sword, in case of several battles in the region.
    /// </summary>
    [JsonPropertyName("battleIndex")]
    public required int BattleIndex { get; set; }
}

/// <summary>
/// Message passed from the client with the ravenMsg when they want to spy on another player's
/// orders. Used for the throne expansion.
/// </summary>
public partial class RavenMessage : GodotObject, ISendableMessage
{
    /// <summary>
    /// ID of the player on whom to spy.
    /// </summary>
    [JsonPropertyName("player")]
    public required string Player { get; set; }
}
