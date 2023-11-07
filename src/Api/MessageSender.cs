using System;
using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Godot;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Api;

internal class MessageSender
{
    public readonly BlockingCollection<GodotObject> SendQueue = new();
    private readonly ClientWebSocket _websocket;

    public MessageSender(ClientWebSocket websocket)
    {
        _websocket = websocket;
    }

    public void SendMessagesFromQueue(CancellationToken cancellationToken)
    {
        while (!SendQueue.IsCompleted)
        {
            if (cancellationToken.IsCancellationRequested)
                return;

            try
            {
                if (_websocket.State != WebSocketState.Open)
                {
                    Task.Delay(50, cancellationToken).GetAwaiter().GetResult();
                    continue;
                }

                var message = SendQueue.Take(cancellationToken);
                var serializedMessage = SerializeMessage(message);
                _websocket
                    .SendAsync(
                        serializedMessage,
                        WebSocketMessageType.Text,
                        true,
                        cancellationToken
                    )
                    .GetAwaiter()
                    .GetResult();
            }
            catch (Exception e)
            {
                // If we were canceled, we don't want to show an error
                if (cancellationToken.IsCancellationRequested)
                    return;

                MessageDisplay.Instance.ShowError("Failed to send message to server", e.Message);
            }
        }
    }

    private static byte[] SerializeMessage(GodotObject message)
    {
        if (
            !MessageDictionary.SendableMessageTypes.TryGetValue(
                message.GetType(),
                out var messageTag
            )
        )
            throw new Exception($"Unrecognized type of message object: '{message.GetType()}'");

        return JsonSerializer.SerializeToUtf8Bytes(
            new Message { Tag = messageTag, Data = message }
        );
    }
}
