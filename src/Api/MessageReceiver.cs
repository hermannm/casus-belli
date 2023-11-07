using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Immerse.BfhClient.Api.Messages;
using Godot;

namespace Immerse.BfhClient.Api;

/// <summary>
/// Handles receiving messages from the WebSocket connection to the game server.
/// </summary>
internal class MessageReceiver
{
    private readonly ClientWebSocket _websocket;

    private readonly Dictionary<MessageType, IMessageReceiveQueue> _messageQueuesByTypeField =
        new();
    private readonly Dictionary<Type, IMessageReceiveQueue> _messageQueuesByActualType = new();

    /// <summary>
    /// Queues where messages are placed as they are received from the server, one for each message
    /// type.
    /// </summary>
    public IEnumerable<IMessageReceiveQueue> MessageQueues => _messageQueuesByTypeField.Values;

    public MessageReceiver(ClientWebSocket websocket)
    {
        _websocket = websocket;
    }

    /// <summary>
    /// Spawns a thread that continuously listens for messages on the WebSocket connection.
    /// Stops the thread when the given cancellation token is canceled.
    /// </summary>
    public void StartReceivingMessages(CancellationToken cancellationToken)
    {
        new Thread(() => ReceiveMessagesIntoQueues(cancellationToken)).Start();
    }

    /// <summary>
    /// Registers the given message type, with the corresponding message ID, as a message that the
    /// client expects to receive from the server.
    /// </summary>
    public void RegisterReceivableMessage<TMessage>(MessageType messageType)
        where TMessage : IReceivableMessage
    {
        var queue = new MessageReceiveQueue<TMessage>();
        _messageQueuesByTypeField.Add(messageType, queue);
        _messageQueuesByActualType.Add(typeof(TMessage), queue);
    }

    /// <summary>
    /// Gets the message queue corresponding to the given message type.
    /// </summary>
    /// <exception cref="ArgumentException">If no queue was found for the given type.</exception>
    public MessageReceiveQueue<TMessage> GetMessageQueueByType<TMessage>()
        where TMessage : IReceivableMessage
    {
        if (!_messageQueuesByActualType.TryGetValue(typeof(TMessage), out var queue))
        {
            throw new ArgumentException($"Unrecognized message type: '{typeof(TMessage)}'");
        }

        return (MessageReceiveQueue<TMessage>)queue;
    }

    /// <summary>
    /// Continuously reads incoming messages from the WebSocket connection.
    /// After a message is read to completion, calls <see cref="DeserializeAndEnqueueMessage"/> to
    /// deserialize and enqueue the message appropriately.
    /// </summary>
    /// <remarks>
    /// Implementation based on https://www.patrykgalach.com/2019/11/11/implementing-websocket-in-unity/.
    /// </remarks>
    private async void ReceiveMessagesIntoQueues(CancellationToken cancellationToken)
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

            var memoryStream = new MemoryStream();
            var isTextMessage = true;

            while (true)
            {
                var buffer = new ArraySegment<byte>(new byte[4 * 1024]);

                var chunkResult = await _websocket.ReceiveAsync(buffer, cancellationToken);
                if (chunkResult.MessageType == WebSocketMessageType.Text)
                {
                    isTextMessage = false;
                    break;
                }

                memoryStream.Write(buffer.Array!, buffer.Offset, chunkResult.Count);

                if (chunkResult.EndOfMessage)
                {
                    break;
                }
            }

            if (!isTextMessage)
            {
                GD.PrintErr("Received unexpected non-text message from WebSocket connection");
                continue;
            }

            memoryStream.Seek(0, SeekOrigin.Begin);

            using var reader = new StreamReader(memoryStream, Encoding.UTF8);
            var messageString = await reader.ReadToEndAsync(cancellationToken);

            try
            {
                DeserializeAndEnqueueMessage(messageString);
            }
            catch (Exception exception)
            {
                GD.PrintErr($"Failed to deserialize received message: {exception.Message}");
            }
        }
    }

    /// <summary>
    /// Messages received from the server are JSON on the following format:
    /// <code>
    /// {
    ///     "[messageId]": {...message}
    /// }
    /// </code>
    /// This method takes the full message JSON string, deserializes the "wrapping object" to get
    /// the message ID, then calls on the appropriate message queue to further deserialize and
    /// enqueue the wrapped message object.
    /// </summary>
    /// <exception cref="ArgumentException">
    /// If no message queue was found for the message's ID.
    /// </exception>
    private void DeserializeAndEnqueueMessage(string messageString)
    {
        var json = JsonDocument.Parse(messageString).RootElement;
        var messageType = json.GetProperty("type").Deserialize<MessageType>();

        if (!_messageQueuesByTypeField.TryGetValue(messageType, out var queue))
        {
            throw new ArgumentException(
                $"Unrecognized message type received from server: '{messageType}'"
            );
        }

        queue.DeserializeAndEnqueueMessage(json.GetProperty("data"));
    }
}
