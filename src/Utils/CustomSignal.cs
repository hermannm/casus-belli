using System;
using Godot;
using GodotArray = Godot.Collections.Array;
using GodotDictionary = Godot.Collections.Dictionary;

namespace Immerse.BfhClient.Utils;

public readonly struct CustomSignal
{
    private readonly StringName _signal;
    private readonly Node _node = new();

    public CustomSignal(StringName name)
    {
        _signal = name;
        _node.RegisterCustomSignal(_signal);
    }

    public void Emit() => _node.EmitCustomSignal(_signal);

    public void AddListener(Action listener) => _node.ConnectCustomSignal(_signal, listener);
}

public readonly struct CustomSignal<T>
{
    private readonly StringName _signal;
    private readonly Node _node = new();

    public CustomSignal(StringName name)
    {
        _signal = name;
        _node.RegisterCustomSignal(_signal);
    }

    public void Emit(T param) => _node.EmitCustomSignal(_signal, param);

    public void AddListener(Action<T> listener) => _node.ConnectCustomSignal(_signal, listener);
}

public readonly struct CustomSignal<T1, T2>
{
    private readonly StringName _signal;
    private readonly Node _node = new();

    public CustomSignal(StringName name)
    {
        _signal = name;
        _node.RegisterCustomSignal(_signal);
    }

    public void Emit(T1 param1, T2 param2) => _node.EmitCustomSignal(_signal, param1, param2);

    public void AddListener(Action<T1, T2> listener) =>
        _node.ConnectCustomSignal(_signal, listener);
}
