using Godot;

namespace CasusBelli.Client.Utils;

public static class ClearChildrenExtension
{
    public static void ClearChildren(this Node node)
    {
        foreach (var child in node.GetChildren())
        {
            child.QueueFree();
        }
    }
}
