using Godot;

namespace Immerse.BfhClient.Utils;

public static class ClearChildrenExtension
{
    public static void ClearChildren(this Node node)
    {
        foreach (var child in node.GetChildren())
        {
            node.RemoveChild(child);
        }
    }
}
