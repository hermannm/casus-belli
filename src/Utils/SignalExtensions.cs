using System;
using Godot;
using GodotArray = Godot.Collections.Array;
using GodotDictionary = Godot.Collections.Dictionary;

namespace Immerse.BfhClient.Utils;

public static class SignalExtensions
{
    public static void ConnectCustomSignal(this Node node, StringName signal, Action listener)
    {
        ConnectCustomSignalInner(node, signal, Callable.From(listener));
    }

    // Overload for Action with one parameter, since Action and Action<T> are not interchangeable.
    public static void ConnectCustomSignal<T>(this Node node, StringName signal, Action<T> listener)
    {
        ConnectCustomSignalInner(node, signal, Callable.From(listener));
    }

    // Overload for Action with two parameters.
    public static void ConnectCustomSignal<T1, T2>(
        this Node node,
        StringName signal,
        Action<T1, T2> listener
    )
    {
        ConnectCustomSignalInner(node, signal, Callable.From(listener));
    }

    private static void ConnectCustomSignalInner(Node node, StringName signal, Callable handler)
    {
        var error = node.Connect(signal, handler);
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to connect signal '{signal}': {error}");
        }
    }

    public static void EmitCustomSignal(this Node node, StringName signal)
    {
        EmitCustomSignalInner(node, signal);
    }

    public static void EmitCustomSignal<T>(this Node node, StringName signal, T param)
    {
        EmitCustomSignalInner(node, signal, Variant.From(param));
    }

    public static void EmitCustomSignal<T1, T2>(
        this Node node,
        StringName signal,
        T1 param1,
        T2 param2
    )
    {
        EmitCustomSignalInner(node, signal, Variant.From(param1), Variant.From(param2));
    }

    private static void EmitCustomSignalInner(Node node, StringName signal, params Variant[] args)
    {
        var error = node.EmitSignal(signal, args);
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
