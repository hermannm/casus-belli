using System;
using Godot;

namespace Immerse.BfhClient.UI;

/// <summary>
/// Utility extension method for getting the global ApiClient instance from any node.
/// The ApiClient should always be available, since it is configured to autoload in Godot.
/// </summary>
public static class MessageDisplayGetterExtension
{
    public static MessageDisplay GetMessageDisplay(this Node node)
    {
        return node.GetNode<MessageDisplay>("/root/MessageDisplay")
            ?? throw new Exception(
                "Failed to find MessageDisplay node - is it added in the project autoload list?"
            );
    }
}
