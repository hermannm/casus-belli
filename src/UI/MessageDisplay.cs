using System;
using System.Text;
using Godot;

namespace Immerse.BfhClient.UI;

public partial class MessageDisplay : Node
{
    /// MessageDisplay singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static MessageDisplay Instance { get; private set; } = null!;

    private VBoxContainer _messageContainer = null!; // Set in _Ready

    public override void _EnterTree()
    {
        Instance = this;
    }

    public override void _Ready()
    {
        _messageContainer = GetNode<VBoxContainer>("%MessageContainer");
    }

    public void ShowError(string errorMessage, params string[] subErrors)
    {
        var errorMessageNode = ResourceLoader.Load<PackedScene>(Scenes.ErrorMessage).Instantiate();
        var label = errorMessageNode.GetNode<Label>("%ErrorMessageLabel");
        var button = errorMessageNode.GetNode<TextureButton>("%CloseErrorButton");

        var stringBuilder = new StringBuilder();
        stringBuilder.Append("Error: ");
        stringBuilder.Append(char.ToUpper(errorMessage[0]));
        stringBuilder.Append(errorMessage.AsSpan(1));
        foreach (var error in subErrors)
        {
            stringBuilder.Append('\n');
            stringBuilder.Append('-');
            stringBuilder.Append(' ');
            stringBuilder.Append(error);
        }

        label.Text = stringBuilder.ToString();
        _messageContainer.CallDeferred("add_child", errorMessageNode);

        button.Pressed += () =>
        {
            _messageContainer.CallDeferred("remove_child", errorMessageNode);
        };
    }
}
