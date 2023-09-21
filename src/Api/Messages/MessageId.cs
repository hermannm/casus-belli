namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// <para>
/// IDs for the types of messages sent between client and server.
/// Each ID corresponds to a message struct in <see cref="Immerse.BfhClient.Api.Messages"/>.
/// </para>
///
/// <para>
/// Message types are used as keys in the JSON messages to and from the server.
/// Every message has the following format, where messageID is one of the <see cref="MessageId"/>
/// constants, and {...message} is the corresponding "...Message" struct in
/// <see cref="Immerse.BfhClient.Api.Messages"/>.
/// <code>
/// {
///     "[messageId]": {...message}
/// }
/// </code>
/// </para>
/// </summary>
///
/// <example>
/// <see cref="MessageId.SupportRequest"/> is the message ID for <see cref="SupportRequestMessage"/>.
/// The message looks like this when coming from the server:
/// <code>
/// {
///     "supportRequest": {
///         "supportingRegion": "Calis",
///         "supportablePlayers": ["red", "green"]
///     }
/// }
/// </code>
/// </example>
public static class MessageId
{
    /// <summary>
    /// Message ID for <see cref="ErrorMessage"/>.
    /// </summary>
    public const string Error = "error";

    /// <summary>
    /// Message ID for <see cref="PlayerStatusMessage"/>.
    /// </summary>
    public const string PlayerStatus = "playerStatus";

    /// <summary>
    /// Message ID for <see cref="LobbyJoinedMessage"/>.
    /// </summary>
    public const string LobbyJoined = "lobbyJoined";

    /// <summary>
    /// Message ID for <see cref="SelectGameIdMessage"/>.
    /// </summary>
    public const string SelectGameId = "selectGameId";

    /// <summary>
    /// Message ID for <see cref="ReadyMessage"/>.
    /// </summary>
    public const string Ready = "ready";

    /// <summary>
    /// Message ID for <see cref="StartGameMessage"/>.
    /// </summary>
    public const string StartGame = "startGame";

    /// <summary>
    /// Message ID for <see cref="SupportRequestMessage"/>.
    /// </summary>
    public const string SupportRequest = "supportRequest";

    /// <summary>
    /// Message ID for <see cref="OrderRequestMessage"/>.
    /// </summary>
    public const string OrderRequest = "orderRequest";

    /// <summary>
    /// Message ID for <see cref="OrdersReceivedMessage"/>.
    /// </summary>
    public const string OrdersReceived = "ordersReceived";

    /// <summary>
    /// Message ID for <see cref="OrdersConfirmationMessage"/>.
    /// </summary>
    public const string OrdersConfirmation = "ordersConfirmation";

    /// <summary>
    /// Message ID for <see cref="BattleResultsMessage"/>.
    /// </summary>
    public const string BattleResults = "battleResults";

    /// <summary>
    /// Message ID for <see cref="WinnerMessage"/>.
    /// </summary>
    public const string Winner = "winner";

    /// <summary>
    /// Message ID for <see cref="SubmitOrdersMessage"/>.
    /// </summary>
    public const string SubmitOrders = "submitOrders";

    /// <summary>
    /// Message ID for <see cref="GiveSupportMessage"/>.
    /// </summary>
    public const string GiveSupport = "giveSupport";

    /// <summary>
    /// Message ID for <see cref="WinterVoteMessage"/>.
    /// </summary>
    public const string WinterVote = "winterVote";

    /// <summary>
    /// Message ID for <see cref="SwordMessage"/>.
    /// </summary>
    public const string Sword = "sword";

    /// <summary>
    /// Message ID for <see cref="RavenMessage"/>.
    /// </summary>
    public const string Raven = "raven";
}
