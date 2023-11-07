using System;
using System.Collections.Generic;
using System.Linq;

namespace Immerse.BfhClient.Api.Messages;

public static class MessageDictionary
{
    public static readonly Dictionary<MessageTag, Type> ReceivableMessageTags =
        new()
        {
            { MessageTag.Error, typeof(ErrorMessage) },
            { MessageTag.PlayerStatus, typeof(PlayerStatusMessage) },
            { MessageTag.LobbyJoined, typeof(LobbyJoinedMessage) },
            { MessageTag.SupportRequest, typeof(SupportRequestMessage) },
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) },
            { MessageTag.OrderRequest, typeof(OrderRequestMessage) },
            { MessageTag.OrdersReceived, typeof(OrdersReceivedMessage) },
            { MessageTag.OrdersConfirmation, typeof(OrdersConfirmationMessage) },
            { MessageTag.BattleResults, typeof(BattleResultsMessage) },
            { MessageTag.Winner, typeof(WinnerMessage) }
        };

    public static readonly Dictionary<MessageTag, Type> SendableMessageTags =
        new()
        {
            { MessageTag.SelectGameId, typeof(SelectGameIdMessage) },
            { MessageTag.Ready, typeof(ReadyMessage) },
            { MessageTag.StartGame, typeof(StartGameMessage) },
            { MessageTag.SubmitOrders, typeof(SubmitOrdersMessage) },
            { MessageTag.GiveSupport, typeof(GiveSupportMessage) },
            { MessageTag.WinterVote, typeof(WinterVoteMessage) },
            { MessageTag.Sword, typeof(SwordMessage) },
            { MessageTag.Raven, typeof(RavenMessage) },
        };

    public static readonly Dictionary<Type, MessageTag> ReceivableMessageTypes =
        ReceivableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);

    public static readonly Dictionary<Type, MessageTag> SendableMessageTypes =
        SendableMessageTags.ToDictionary(pair => pair.Value, pair => pair.Key);
}
