using System;
using Godot;

namespace Immerse.BfhClient.Utils;

public static class ConnectSignalExtension
{
    public static void ConnectSignal(this Node node, StringName signal, Action signalHandler)
    {
        var error = node.Connect(signal, Callable.From(signalHandler));
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to connect signal '{signal}': {error}");
        }
    }

    // Overload for Action with one parameter, since Action and Action<T> are not interchangeable.
    public static void ConnectSignal<T>(this Node node, StringName signal, Action<T> signalHandler)
    {
        var error = node.Connect(signal, Callable.From(signalHandler));
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to connect signal '{signal}': {error}");
        }
    }
}
