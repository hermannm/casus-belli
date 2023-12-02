using System;
using System.Collections.Generic;
using System.Linq;

namespace CasusBelli.Client.Api;

public enum MessageTag
{
    Error = 1,
    PlayerStatus,
    LobbyJoined,
    SelectFaction,
    StartGame,
    GameStarted,
    OrderRequest,
    OrdersReceived,
    OrdersConfirmation,
    BattleAnnouncement,
    BattleResults,
    Winner,
    SubmitOrders,
    GiveSupport,
    DiceRoll
}

public static class MessageTagMap
{
    public static readonly Dictionary<MessageTag, Type> ReceivableMessageTags =
        new()
        {
            { MessageTag.Error, typeof(ErrorMessage) },
            { MessageTag.GameStarted, typeof(GameStartedMessage) },
            { MessageTag.PlayerStatus, typeof(PlayerStatusMessage) },
            { MessageTag.LobbyJoined, typeof(LobbyJoinedMessage) },
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) },
            { MessageTag.OrderRequest, typeof(OrderRequestMessage) },
            { MessageTag.OrdersReceived, typeof(OrdersReceivedMessage) },
            { MessageTag.OrdersConfirmation, typeof(OrdersConfirmationMessage) },
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
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) },
            { MessageTag.DiceRoll, typeof(DiceRollMessage) }
        };

    public static readonly Dictionary<Type, MessageTag> ReceivableMessageTypes =
        ReceivableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);

    public static readonly Dictionary<Type, MessageTag> SendableMessageTypes =
        SendableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);
}
