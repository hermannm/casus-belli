using System;
using Godot;
using Immerse.BfhClient.Api.Messages;

namespace Immerse.BfhClient.Api;

public static class MessageListenerExtension
{
    public static void AddServerMessageHandler<TMessage>(this Node node, Action<TMessage> handler)
        where TMessage : GodotObject, IReceivableMessage
    {
        if (
            !MessageDictionary.ReceivableMessageTypes.TryGetValue(
                typeof(TMessage),
                out var messageTag
            )
        )
        {
            GD.PushError($"Invalid message type {typeof(TMessage)} for for server message handler");
            return;
        }

        var signal = ApiClient.GetMessageReceivedSignalName(messageTag);
        var err = node.Connect(signal, Callable.From(handler));
        if (err != Error.Ok)
            GD.PushError($"Failed to connect to signal '{signal}': {err}");
    }
}
