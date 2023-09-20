using System.Collections.Generic;
using Immerse.BfhClient.Api.GameTypes;
using Newtonsoft.Json;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Message sent from server when asking a supporting player who to support in an embattled region.
/// </summary>
public struct SupportRequestMessage : IReceivableMessage
{
    /// <summary>
    /// The region from which support is asked, where the asked player should have a support order.
    /// </summary>
    [JsonProperty("supportingRegion")]
    public string SupportingRegion;

    /// <summary>
    /// List of possible players to support in the battle.
    /// </summary>
    [JsonProperty("supportablePlayers")]
    public List<string> SupportablePlayers;
}

/// <summary>
/// Message sent from server to client to signal that client should submit orders.
/// </summary>
public struct OrderRequestMessage : IReceivableMessage { }

/// <summary>
/// Message sent from server to all clients when valid orders are received from all players.
/// </summary>
public struct OrdersReceivedMessage : IReceivableMessage
{
    /// <summary>
    /// Maps a player's ID to their submitted orders.
    /// </summary>
    [JsonProperty("playerOrders")]
    public Dictionary<string, List<Order>> PlayerOrders;
}

/// <summary>
/// Message sent from server to all clients when valid orders are received from a player.
/// Used to show who the server is waiting for.
/// </summary>
public struct OrdersConfirmationMessage : IReceivableMessage
{
    /// <summary>
    /// The player who submitted orders.
    /// </summary>
    [JsonProperty("player")]
    public string Player;
}

/// <summary>
/// Message sent from server to all clients when a battle result is calculated.
/// </summary>
public struct BattleResultsMessage : IReceivableMessage
{
    /// <summary>
    /// The relevant battle result.
    /// </summary>
    [JsonProperty("battles")]
    public List<Battle> Battles;
}

/// <summary>
/// Message sent from server to all clients when the game is won.
/// </summary>
public struct WinnerMessage : IReceivableMessage
{
    /// <summary>
    /// Player tag of the game's winner.
    /// </summary>
    [JsonProperty("winner")]
    public string Winner;
}

/// <summary>
/// Message sent from client when submitting orders.
/// </summary>
public struct SubmitOrdersMessage : ISendableMessage
{
    /// <summary>
    /// List of submitted orders.
    /// </summary>
    [JsonProperty("orders")]
    public List<Order> Orders;
}

/// <summary>
/// Message sent from client when declaring who to support with their support order.
/// Forwarded by server to all clients to show who were given support.
/// </summary>
public struct GiveSupportMessage : IReceivableMessage, ISendableMessage
{
    /// <summary>
    /// Name of the region in which the support order is placed.
    /// </summary>
    [JsonProperty("supportingRegion")]
    public string SupportingRegion;

    /// <summary>
    /// ID of the player in the destination region to support.
    /// Null if none were supported.
    /// </summary>
    [JsonProperty("supportedPlayer")]
    public string? SupportedPlayer;
}

/// <summary>
/// Message passed from the client during winter council voting.
/// Used for the throne expansion.
/// </summary>
public struct WinterVoteMessage : ISendableMessage
{
    /// <summary>
    /// ID of the player that the submitting player votes for.
    /// </summary>
    [JsonProperty("player")]
    public string Player;
}

/// <summary>
/// Message passed from the client with the swordMsg to declare where they want to use it.
/// Used for the throne expansion.
/// </summary>
public struct SwordMessage : ISendableMessage
{
    /// <summary>
    /// Name of the region in which the player wants to use the sword in battle.
    /// </summary>
    [JsonProperty("region")]
    public string Region;

    /// <summary>
    /// Index of the battle in which to use the sword, in case of several battles in the region.
    /// </summary>
    [JsonProperty("battleIndex")]
    public int BattleIndex;
}

/// <summary>
/// Message passed from the client with the ravenMsg when they want to spy on another player's orders.
/// Used for the throne expansion.
/// </summary>
public struct RavenMessage : ISendableMessage
{
    /// <summary>
    /// ID of the player on whom to spy.
    /// </summary>
    [JsonProperty("player")]
    public string Player;
}
