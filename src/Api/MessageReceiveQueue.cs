using System;
using System.Collections.Concurrent;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Immerse.BfhClient.Api.Messages;

namespace Immerse.BfhClient.Api;

/// <summary>
/// Utility interface to enable keeping <see cref="MessageReceiveQueue{TMessage}"/>s of different types in a
/// collection.
/// </summary>
internal interface IMessageReceiveQueue
{
    public Task CheckReceivedMessages(CancellationToken cancellationToken);
    public void DeserializeAndEnqueueMessage(JsonElement serializedMessage);
}

/// <summary>
/// Provides a thread-safe queue for messages received from the server, and an event that is triggered when a
/// message is received.
/// </summary>
internal class MessageReceiveQueue<TMessage> : IMessageReceiveQueue
    where TMessage : IReceivableMessage
{
    public event Action<TMessage>? ReceivedMessage;

    private readonly ConcurrentQueue<TMessage> _queue = new();

    /// <summary>
    /// Continuously checks for received messages on the queue.
    /// When a message is received, calls all subscribers to <see cref="ReceivedMessage"/> with the
    /// message.
    /// </summary>
    public async Task CheckReceivedMessages(CancellationToken cancellationToken)
    {
        while (true)
        {
            if (cancellationToken.IsCancellationRequested)
                return;

            if (_queue.TryDequeue(out var message))
            {
                ReceivedMessage?.Invoke(message);
            }

            await Task.Yield();
        }
    }

    /// <summary>
    /// Attempts to deserialize the given message to the message type held by this queue.
    /// If deserialization succeeds, adds the message to the queue.
    /// </summary>
    /// <exception cref="ArgumentException">
    /// If the given message could not be deserialized to message type of the queue.
    /// </exception>
    public void DeserializeAndEnqueueMessage(JsonElement serializedMessage)
    {
        var message =
            serializedMessage.Deserialize<TMessage>()
            ?? throw new ArgumentException($"Failed to deserialize message '{serializedMessage}'");

        _queue.Enqueue(message);
    }
}
