using System;
using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Text.Json;
using System.Text.Json.Serialization;
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
            {
                return;
            }

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
                        new CancellationToken()
                    )
                    .GetAwaiter()
                    .GetResult();
            }
            catch (Exception e)
            {
                // If we were canceled, we don't want to show an error
                if (cancellationToken.IsCancellationRequested)
                {
                    return;
                }

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
        {
            throw new Exception($"Unrecognized type of message object: '{message.GetType()}'");
        }

        var options = new JsonSerializerOptions();
        options.Converters.Add(new MessageDataSerializer());

        return JsonSerializer.SerializeToUtf8Bytes(
            new Message { Tag = messageTag, Data = message },
            options
        );
    }

    /// <summary>
    /// Custom serializer to avoid serializing IntPtr fields from GodotObject, which causes
    /// serialization to fail.
    /// </summary>
    private class MessageDataSerializer : JsonConverter<GodotObject>
    {
        public override void Write(
            Utf8JsonWriter writer,
            GodotObject messageData,
            JsonSerializerOptions options
        )
        {
            writer.WriteStartObject();

            foreach (var property in messageData.GetType().GetProperties())
            {
                if (property.PropertyType == typeof(IntPtr))
                {
                    continue;
                }

                var propValue = property.GetValue(messageData);
                if (
                    propValue is not null
                    || (
                        propValue is null
                        && options.DefaultIgnoreCondition == JsonIgnoreCondition.Never
                    )
                )
                {
                    writer.WritePropertyName(property.Name);
                    JsonSerializer.Serialize(writer, propValue, options);
                    break;
                }
            }

            writer.WriteEndObject();
        }

        public override GodotObject Read(
            ref Utf8JsonReader reader,
            Type typeToConvert,
            JsonSerializerOptions options
        )
        {
            throw new NotImplementedException();
        }
    }
}
