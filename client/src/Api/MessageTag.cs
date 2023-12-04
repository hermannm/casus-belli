using System;
using System.Collections.Generic;
using System.Linq;

namespace CasusBelli.Client.Api;

public enum MessageTag
{
    Error = 1,
    LobbyJoined,
    PlayerStatus,
    SelectFaction,
    StartGame,
    GameStarted,
    OrderRequest,
    OrdersConfirmation,
    OrdersReceived,
    BattleAnnouncement,
    BattleResults,
    Winner,
    SubmitOrders,
    DiceRoll,
    GiveSupport
}

public static class MessageTagMap
{
    public static readonly Dictionary<MessageTag, Type> ReceivableMessageTags =
        new()
        {
            { MessageTag.Error, typeof(ErrorMessage) },
            { MessageTag.LobbyJoined, typeof(LobbyJoinedMessage) },
            { MessageTag.PlayerStatus, typeof(PlayerStatusMessage) },
            { MessageTag.GameStarted, typeof(GameStartedMessage) },
            { MessageTag.OrderRequest, typeof(OrderRequestMessage) },
            { MessageTag.OrdersConfirmation, typeof(OrdersConfirmationMessage) },
            { MessageTag.OrdersReceived, typeof(OrdersReceivedMessage) },
            { MessageTag.BattleAnnouncement, typeof(BattleAnnouncementMessage) },
            { MessageTag.BattleResults, typeof(BattleResultsMessage) },
            { MessageTag.Winner, typeof(WinnerMessage) }
        };

    public static readonly Dictionary<MessageTag, Type> SendableMessageTags =
        new()
        {
            { MessageTag.SelectFaction, typeof(SelectFactionMessage) },
            { MessageTag.StartGame, typeof(StartGameMessage) },
            { MessageTag.SubmitOrders, typeof(SubmitOrdersMessage) },
            { MessageTag.DiceRoll, typeof(DiceRollMessage) },
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) }
        };

    public static readonly Dictionary<Type, MessageTag> ReceivableMessageTypes =
        ReceivableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);

    public static readonly Dictionary<Type, MessageTag> SendableMessageTypes =
        SendableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);
}
