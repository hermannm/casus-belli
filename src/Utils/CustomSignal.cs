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

    public void Emit()
    {
        _node.EmitCustomSignal(_signal);
    }

    public void Connect(Action callback)
    {
        _node.ConnectCustomSignal(_signal, callback);
    }
}

public readonly struct CustomSignal<[MustBeVariant] T>
{
    private readonly StringName _signal;
    private readonly Node _node = new();

    public CustomSignal(StringName name)
    {
        _signal = name;
        _node.RegisterCustomSignal(_signal);
    }

    public void Emit(T param)
    {
        _node.EmitCustomSignal(_signal, param);
    }

    public void Connect(Action<T> callback)
    {
        _node.ConnectCustomSignal(_signal, callback);
    }
}
