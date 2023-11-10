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
    [JsonPropertyName("supportingRegion")]
    public required string SupportingRegion { get; set; }

    [JsonPropertyName("embattledRegion")]
    public required string EmbattledRegion { get; set; }

    [JsonPropertyName("supportableFactions")]
    public required List<string> SupportableFactions { get; set; }
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
    [JsonPropertyName("ordersByFaction")]
    public required Dictionary<string, List<Order>> OrdersByFaction { get; set; }
}

/// <summary>
/// Message sent from server to all clients when valid orders are received from a player.
/// Used to show who the server is waiting for.
/// </summary>
public partial class OrdersConfirmationMessage : GodotObject, IReceivableMessage
{
    [JsonPropertyName("factionThatSubmittedOrders")]
    public required string FactionThatSubmittedOrders { get; set; }
}

/// <summary>
/// Message sent from server to all clients when a battle result is calculated.
/// </summary>
public partial class BattleResultsMessage : GodotObject, IReceivableMessage
{
    [JsonPropertyName("battles")]
    public required List<Battle> Battles { get; set; }
}

/// <summary>
/// Message sent from server to all clients when the game is won.
/// </summary>
public partial class WinnerMessage : GodotObject, IReceivableMessage
{
    [JsonPropertyName("winningFaction")]
    public required string WinningFaction { get; set; }
}

/// <summary>
/// Message sent from client when submitting orders.
/// </summary>
public partial class SubmitOrdersMessage : GodotObject, ISendableMessage
{
    [JsonPropertyName("orders")]
    public required List<Order> Orders { get; set; }
}

/// <summary>
/// Message sent from client when declaring who to support with their support order.
/// Forwarded by server to all clients to show who were given support.
/// </summary>
public partial class GiveSupportMessage : GodotObject, IReceivableMessage, ISendableMessage
{
    [JsonPropertyName("supportingRegion")]
    public required string SupportingRegion { get; set; }

    [JsonPropertyName("embattledRegion")]
    public required string EmbattledRegion { get; set; }

    /// <summary>
    /// Null if none were supported.
    /// </summary>
    [JsonPropertyName("supportedFaction")]
    public string? SupportedFaction { get; set; }
}

/// <summary>
/// Message passed sent from the client during winter council voting.
/// Used for the throne expansion.
/// </summary>
public partial class WinterVoteMessage : GodotObject, ISendableMessage
{
    [JsonPropertyName("factionVotedFor")]
    public required string FactionVotedFor { get; set; }
}

/// <summary>
/// Message sent from the client with the sword to declare where they want to use it.
/// Used for the throne expansion.
/// </summary>
public partial class SwordMessage : GodotObject, ISendableMessage
{
    [JsonPropertyName("region")]
    public required string Region { get; set; }

    /// <summary>
    /// Index of the battle in which to use the sword, in case of several battles in the region.
    /// </summary>
    [JsonPropertyName("battleIndex")]
    public required int BattleIndex { get; set; }
}

/// <summary>
/// Message sent from the client with the ravenMsg when they want to spy on another player's
/// orders. Used for the throne expansion.
/// </summary>
public partial class RavenMessage : GodotObject, ISendableMessage
{
    [JsonPropertyName("factionToSpyOn")]
    public required string FactionToSpyOn { get; set; }
}
