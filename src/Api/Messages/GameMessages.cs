using System.Collections.Generic;
using Godot;
using Immerse.BfhClient.Api.GameTypes;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Message sent from server when asking a supporting player who to support in an embattled region.
/// </summary>
public partial class SupportRequestMessage : GodotObject, IReceivableMessage
{
    public required string SupportingRegion { get; set; }
    public required string EmbattledRegion { get; set; }
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
    public required Dictionary<string, List<Order>> OrdersByFaction { get; set; }
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
/// Message sent from server to all clients when a battle result is calculated.
/// </summary>
public partial class BattleResultsMessage : GodotObject, IReceivableMessage
{
    public required List<Battle> Battles { get; set; }
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
    public required string SupportingRegion { get; set; }
    public required string EmbattledRegion { get; set; }

    /// <summary>
    /// Null if none were supported.
    /// </summary>
    public string? SupportedFaction { get; set; }
}
