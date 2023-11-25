namespace CasusBelli.Client.Api.Messages;

public enum MessageTag
{
    Error = 1,
    PlayerStatus,
    LobbyJoined,
    SelectFaction,
    StartGame,
    GameStarted,
    SupportRequest,
    OrderRequest,
    OrdersReceived,
    OrdersConfirmation,
    BattleResults,
    DangerZoneCrossings,
    Winner,
    SubmitOrders,
    GiveSupport
}
