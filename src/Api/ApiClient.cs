using System;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;
using Immerse.BfhClient.Api.Messages;
using Godot;

namespace Immerse.BfhClient.Api;

/// <summary>
/// WebSocket client that connects to the game server.
/// Provides methods for sending and receiving messages to and from the server.
/// </summary>
// ReSharper disable once ClassNeverInstantiated.Global
public partial class ApiClient : Node
{
    private readonly ClientWebSocket _connection;
    private readonly MessageSender _messageSender;
    private readonly MessageReceiver _messageReceiver;
    private readonly CancellationTokenSource _cancellation;

    public ApiClient()
    {
        _connection = new ClientWebSocket();
        _messageSender = new MessageSender(_connection);
        _messageReceiver = new MessageReceiver(_connection);
        _cancellation = new CancellationTokenSource();

        RegisterSendableMessages();
        RegisterReceivableMessages();
    }

    /// <summary>
    /// Connects the API client to a server at the given URI, and starts sending and receiving
    /// messages.
    /// </summary>
    public Task Connect(Uri serverUri)
    {
        foreach (var messageQueue in _messageReceiver.MessageQueues)
        {
            Task.Run(() => messageQueue.CheckReceivedMessages(_cancellation.Token));
        }

        _messageReceiver.StartReceivingMessages(_cancellation.Token);
        _messageSender.StartSendingMessages(_cancellation.Token);

        return _connection.ConnectAsync(serverUri, _cancellation.Token);
    }

    /// <summary>
    /// Disconnects the API client from the server, and stops sending and receiving messages.
    /// </summary>
    public Task Disconnect()
    {
        _cancellation.Cancel();

        return _connection.CloseAsync(
            WebSocketCloseStatus.NormalClosure,
            "Client initiated disconnect from game server",
            _cancellation.Token
        );
    }

    /// <summary>
    /// Sends the given message to the server.
    /// </summary>
    ///
    /// <typeparam name="TMessage">
    /// Must be registered in <see cref="RegisterSendableMessages"/>, which should be all message
    /// types marked with <see cref="ISendableMessage"/>.
    /// </typeparam>
    public void SendServerMessage<TMessage>(TMessage message)
        where TMessage : ISendableMessage
    {
        _messageSender.SendQueue.Add(message);
    }

    /// <summary>
    /// Registers the given method to be called whenever the server sends a message of the given
    /// type.
    /// </summary>
    ///
    /// <typeparam name="TMessage">
    /// Must be registered in <see cref="RegisterReceivableMessages"/>, which should be all message
    /// types marked with <see cref="IReceivableMessage"/>.
    /// </typeparam>
    public void RegisterServerMessageHandler<TMessage>(Action<TMessage> messageHandler)
        where TMessage : IReceivableMessage
    {
        var queue = _messageReceiver.GetMessageQueueByType<TMessage>();
        queue.ReceivedMessage += messageHandler;
    }

    /// <summary>
    /// Deregisters the given message handler method. Should be called when a message handler is
    /// disposed, to properly remove all references to it.
    /// </summary>
    ///
    /// <typeparam name="TMessage">
    /// Must be registered in <see cref="RegisterReceivableMessages"/>, which should be all message
    /// types marked with <see cref="IReceivableMessage"/>.
    /// </typeparam>
    public void DeregisterServerMessageHandler<TMessage>(Action<TMessage> messageHandler)
        where TMessage : IReceivableMessage
    {
        var queue = _messageReceiver.GetMessageQueueByType<TMessage>();
        queue.ReceivedMessage -= messageHandler;
    }

    /// <summary>
    /// Registers all message types that the client expects to be able to send to the server.
    /// </summary>
    private void RegisterSendableMessages()
    {
        _messageSender.RegisterSendableMessage<SelectGameIdMessage>(MessageId.SelectGameId);
        _messageSender.RegisterSendableMessage<ReadyMessage>(MessageId.Ready);
        _messageSender.RegisterSendableMessage<StartGameMessage>(MessageId.StartGame);
        _messageSender.RegisterSendableMessage<SubmitOrdersMessage>(MessageId.SubmitOrders);
        _messageSender.RegisterSendableMessage<GiveSupportMessage>(MessageId.GiveSupport);
        _messageSender.RegisterSendableMessage<WinterVoteMessage>(MessageId.WinterVote);
        _messageSender.RegisterSendableMessage<SwordMessage>(MessageId.Sword);
        _messageSender.RegisterSendableMessage<RavenMessage>(MessageId.Raven);
    }

    /// <summary>
    /// Registers all message types that the client expects to receive from the server.
    /// </summary>
    private void RegisterReceivableMessages()
    {
        _messageReceiver.RegisterReceivableMessage<ErrorMessage>(MessageId.Error);
        _messageReceiver.RegisterReceivableMessage<PlayerStatusMessage>(MessageId.PlayerStatus);
        _messageReceiver.RegisterReceivableMessage<LobbyJoinedMessage>(MessageId.LobbyJoined);
        _messageReceiver.RegisterReceivableMessage<SupportRequestMessage>(MessageId.SupportRequest);
        _messageReceiver.RegisterReceivableMessage<GiveSupportMessage>(MessageId.GiveSupport);
        _messageReceiver.RegisterReceivableMessage<OrderRequestMessage>(MessageId.OrderRequest);
        _messageReceiver.RegisterReceivableMessage<OrdersReceivedMessage>(MessageId.OrdersReceived);
        _messageReceiver.RegisterReceivableMessage<OrdersConfirmationMessage>(
            MessageId.OrdersConfirmation
        );
        _messageReceiver.RegisterReceivableMessage<BattleResultsMessage>(MessageId.BattleResults);
        _messageReceiver.RegisterReceivableMessage<WinnerMessage>(MessageId.Winner);
    }
}
