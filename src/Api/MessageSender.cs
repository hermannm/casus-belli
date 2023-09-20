using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Godot;
using Immerse.BfhClient.Api.Messages;
using Newtonsoft.Json.Linq;

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

    private readonly ClientWebSocket _connection;
    private Thread? _sendThread;

    private readonly Dictionary<Type, string> _messageIdMap = new();

    public MessageSender(ClientWebSocket connection)
    {
        _connection = connection;
    }

    /// <summary>
    /// Spawns a thread that continuously listens for messages on the WebSocket connection.
    /// </summary>
    public void StartSendingMessages()
    {
        _sendThread = new Thread(SendMessagesFromQueue);
    }

    /// <summary>
    /// Aborts the message sending thread.
    /// </summary>
    public void StopSendingMessages()
    {
        _sendThread?.Abort();
        _sendThread = null;
    }

    /// <summary>
    /// Registers the given message type, with the corresponding message ID, as a message that the
    /// client expects to be able to send to the server.
    /// </summary>
    public void RegisterSendableMessage<TMessage>(string messageId)
        where TMessage : ISendableMessage
    {
        _messageIdMap.Add(typeof(TMessage), messageId);
    }

    /// <summary>
    /// Continuously takes messages from the send queue, serializes them and sends them to the
    /// server.
    /// </summary>
    /// <remarks>
    /// Implementation based on https://www.patrykgalach.com/2019/11/11/implementing-websocket-in-unity/.
    /// </remarks>
    private async void SendMessagesFromQueue()
    {
        while (true)
        {
            if (_connection.State != WebSocketState.Open)
            {
                Task.Delay(50).Wait();
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

                await _connection.SendAsync(
                    serializedMessage,
                    WebSocketMessageType.Text,
                    true,
                    CancellationToken.None
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
        if (!_messageIdMap.TryGetValue(message.GetType(), out var messageId))
        {
            throw new ArgumentException(
                $"Unrecognized type of message object: '{message.GetType()}'"
            );
        }

        var messageJson = new JObject(new JProperty(messageId, message));
        var messageString = messageJson.ToString();
        var messageBytes = Encoding.UTF8.GetBytes(messageString);
        return messageBytes;
    }
}
