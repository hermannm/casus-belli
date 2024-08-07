using System;
using System.Collections.Concurrent;
using System.IO;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using CasusBelli.Client.UI;
using Godot;

namespace CasusBelli.Client.Api;

internal class MessageReceiver
{
    public ConcurrentQueue<Message> MessageQueue { get; } = new();

    public void ClearQueue()
    {
        MessageQueue.Clear();
    }

    public void ReadMessagesIntoQueue(ClientWebSocket socket, CancellationToken cancellationToken)
    {
        while (true)
        {
            if (cancellationToken.IsCancellationRequested)
            {
                return;
            }

            try
            {
                if (socket.State != WebSocketState.Open)
                {
                    Task.Delay(50, cancellationToken).GetAwaiter().GetResult();
                    continue;
                }

                var messageString = ReadMessageFromSocket(socket);
                var message = DeserializeMessage(messageString);
                MessageQueue.Enqueue(message);
            }
            catch (Exception e)
            {
                // If we were canceled, we don't want to show an error
                if (cancellationToken.IsCancellationRequested)
                {
                    return;
                }

                MessageDisplay.Instance.ShowError("Failed to read message from server", e.Message);
            }
        }
    }

    private static string ReadMessageFromSocket(ClientWebSocket socket)
    {
        var memoryStream = new MemoryStream();

        while (true)
        {
            var buffer = new ArraySegment<byte>(new byte[4 * 1024]);

            var chunkResult = socket
                .ReceiveAsync(buffer, new CancellationToken())
                .GetAwaiter()
                .GetResult();
            if (chunkResult.MessageType != WebSocketMessageType.Text)
            {
                throw new Exception("Received non-text message");
            }

            memoryStream.Write(buffer.Array!, buffer.Offset, chunkResult.Count);

            if (chunkResult.EndOfMessage)
            {
                break;
            }
        }

        memoryStream.Seek(0, SeekOrigin.Begin);

        using var reader = new StreamReader(memoryStream, Encoding.UTF8);
        return reader.ReadToEnd();
    }

    private static Message DeserializeMessage(string messageString)
    {
        var json = JsonDocument.Parse(messageString).RootElement;
        var messageTag = json.GetProperty("Tag").Deserialize<MessageTag>();

        if (!MessageTagMap.ReceivableMessageTags.TryGetValue(messageTag, out var messageType))
        {
            throw new Exception($"Unrecognized message type '{messageTag}' received from server");
        }

        var messageData = json.GetProperty("Data").Deserialize(messageType);
        if (messageData is null)
        {
            throw new Exception("Failed to deserialize message");
        }

        return new Message { Tag = messageTag, Data = (GodotObject)messageData };
    }
}
