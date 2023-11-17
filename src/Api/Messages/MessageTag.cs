namespace Immerse.BfhClient.Api.Messages;

public enum MessageTag
{
    Error = 1,
    PlayerStatus,
    LobbyJoined,
    SelectFaction,
    StartGame,
    SupportRequest,
    OrderRequest,
    OrdersReceived,
    OrdersConfirmation,
    BattleResults,
    Winner,
    SubmitOrders,
    GiveSupport
}
