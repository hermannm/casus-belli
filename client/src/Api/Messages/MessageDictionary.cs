using System;
using System.Collections.Generic;
using System.Linq;

namespace CasusBelli.Client.Api.Messages;

public static class MessageDictionary
{
    public static readonly Dictionary<MessageTag, Type> ReceivableMessageTags =
        new()
        {
            { MessageTag.Error, typeof(ErrorMessage) },
            { MessageTag.GameStarted, typeof(GameStartedMessage) },
            { MessageTag.PlayerStatus, typeof(PlayerStatusMessage) },
            { MessageTag.LobbyJoined, typeof(LobbyJoinedMessage) },
            { MessageTag.SupportRequest, typeof(SupportRequestMessage) },
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) },
            { MessageTag.OrderRequest, typeof(OrderRequestMessage) },
            { MessageTag.OrdersReceived, typeof(OrdersReceivedMessage) },
            { MessageTag.OrdersConfirmation, typeof(OrdersConfirmationMessage) },
            { MessageTag.BattleResults, typeof(BattleResultsMessage) },
            { MessageTag.DangerZoneCrossings, typeof(DangerZoneCrossingsMessage) },
            { MessageTag.Winner, typeof(WinnerMessage) }
        };

    public static readonly Dictionary<MessageTag, Type> SendableMessageTags =
        new()
        {
            { MessageTag.SelectFaction, typeof(SelectFactionMessage) },
            { MessageTag.StartGame, typeof(StartGameMessage) },
            { MessageTag.SubmitOrders, typeof(SubmitOrdersMessage) },
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) }
        };

    public static readonly Dictionary<Type, MessageTag> ReceivableMessageTypes =
        ReceivableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);

    public static readonly Dictionary<Type, MessageTag> SendableMessageTypes =
        SendableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);
}