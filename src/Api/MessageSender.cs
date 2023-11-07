using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Godot;
using Immerse.BfhClient.Api.Messages;

namespace Immerse.BfhClient.Api;

/// <summary>
/// Handles sending messages through the WebSocket connection to the game server.
/// </summary>
internal class MessageSender
{
    /// <summary>
    /// Thread-safe queue to place messages, which will be picked up by the send thread and sent to
    /// the server.
    /// </summary>
    public readonly BlockingCollection<ISendableMessage> SendQueue = new();

    private readonly ClientWebSocket _websocket;

    private readonly Dictionary<Type, MessageType> _messageTypeMap = new();

    public MessageSender(ClientWebSocket websocket)
    {
        _websocket = websocket;
    }

    /// <summary>
    /// Spawns a thread that continuously listens for messages on the WebSocket connection.
    /// Stops the thread when the given cancellation token is canceled.
    /// </summary>
    public void StartSendingMessages(CancellationToken cancellationToken)
    {
        new Thread(() => SendMessagesFromQueue(cancellationToken)).Start();
    }

    /// <summary>
    /// Registers the given message type, with the corresponding message ID, as a message that the
    /// client expects to be able to send to the server.
    /// </summary>
    public void RegisterSendableMessage<TMessage>(MessageType messageType)
        where TMessage : ISendableMessage
    {
        _messageTypeMap.Add(typeof(TMessage), messageType);
    }

    /// <summary>
    /// Continuously takes messages from the send queue, serializes them and sends them to the
    /// server.
    /// </summary>
    /// <remarks>
    /// Implementation based on https://www.patrykgalach.com/2019/11/11/implementing-websocket-in-unity/.
    /// </remarks>
    private async void SendMessagesFromQueue(CancellationToken cancellationToken)
    {
        while (true)
        {
            if (cancellationToken.IsCancellationRequested)
                return;

            if (_websocket.State != WebSocketState.Open)
            {
                await Task.Delay(50, cancellationToken).WaitAsync(cancellationToken);
                continue;
            }

            while (!SendQueue.IsCompleted)
            {
                var message = SendQueue.Take();

                byte[] serializedMessage;
                try
                {
                    serializedMessage = SerializeToJson(message);
                }
                catch (Exception exception)
                {
                    GD.PrintErr($"Failed to serialize sent message: {exception.Message}");
                    continue;
                }

                await _websocket.SendAsync(
                    serializedMessage,
                    WebSocketMessageType.Text,
                    true,
                    cancellationToken
                );
            }
        }
    }

    /// <summary>
    /// Serializes the given message to JSON, wrapping it with the appropriate message ID according
    /// to its type.
    /// </summary>
    /// <exception cref="ArgumentException">
    /// If a message ID could not be found for the type of the message. Likely because the message
    /// type has not been registered with <see cref="RegisterSendableMessage{TMessage}"/>.
    /// </exception>
    private byte[] SerializeToJson(ISendableMessage message)
    {
        if (!_messageTypeMap.TryGetValue(message.GetType(), out var messageType))
        {
            throw new ArgumentException(
                $"Unrecognized type of message object: '{message.GetType()}'"
            );
        }

        return JsonSerializer.SerializeToUtf8Bytes(
            new Message { Type = messageType, Data = message }
        );
    }
}
