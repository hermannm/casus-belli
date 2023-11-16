using Godot;

namespace Immerse.BfhClient.Utils;

/// <summary>
/// Godot uses StringNames for strings internally. Therefore, passing a C# string requires
/// converting it to a StringName. This class holds StringName instantiations for commonly used
/// strings, to avoid this conversion cost.
/// </summary>
public static class Strings
{
    public static readonly StringName AddChild = new("add_child");
    public static readonly StringName RemoveChild = new("remove_child");
}
