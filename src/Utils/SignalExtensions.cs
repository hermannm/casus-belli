using System;
using Godot;
using GodotArray = Godot.Collections.Array;
using GodotDictionary = Godot.Collections.Dictionary;

namespace Immerse.BfhClient.Utils;

public static class SignalExtensions
{
    public static void ConnectCustomSignal(this Node node, StringName signal, Action callback)
    {
        var error = node.Connect(signal, Callable.From(callback));
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to connect signal '{signal}': {error}");
        }
    }

    // Overload for Action with one parameter, since Action and Action<T> are not interchangeable.
    public static void ConnectCustomSignal<T>(this Node node, StringName signal, Action<T> callback)
    {
        var error = node.Connect(signal, Callable.From(callback));
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to connect signal '{signal}': {error}");
        }
    }

    public static void EmitCustomSignal(this Node node, StringName signal)
    {
        var error = node.EmitSignal(signal);
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to emit signal '{signal}': {error}");
        }
    }

    public static void EmitCustomSignal<T>(this Node node, StringName signal, T param)
    {
        var error = node.EmitSignal(signal, Variant.From(param));
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to emit signal '{signal}': {error}");
        }
    }

    public static void RegisterCustomSignal(this Node node, StringName name, params Type[] args)
    {
        if (args.Length == 0)
        {
            node.AddUserSignal(name);
            return;
        }

        var godotArgs = new GodotArray();
        var argNumber = 1;
        foreach (var arg in args)
        {
            var godotType = Type.GetTypeCode(arg) switch
            {
                TypeCode.String => Variant.Type.String,
                TypeCode.Boolean => Variant.Type.Bool,
                TypeCode.Int32 => Variant.Type.Int,
                TypeCode.Object => Variant.Type.Object,
                _ => throw new ArgumentException($"Invalid signal argument type '{arg}'")
            };

            godotArgs.Add(
                new GodotDictionary { { "name", $"arg{argNumber}" }, { "type", (int)godotType } }
            );
            argNumber++;
        }

        node.AddUserSignal(name, godotArgs);
    }
}
