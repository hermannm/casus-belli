namespace Immerse.BfhClient.Api.Messages;

public enum MessageTag : byte
{
    Error = 1,
    PlayerStatus,
    LobbyJoined,
    SelectGameId,
    Ready,
    StartGame,
    SupportRequest,
    OrderRequest,
    OrdersReceived,
    OrdersConfirmation,
    BattleResults,
    Winner,
    SubmitOrders,
    GiveSupport,
    WinterVote,
    Sword,
    Raven
}
