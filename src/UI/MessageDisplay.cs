using System;
using System.Text;
using Godot;

namespace Immerse.BfhClient.UI;

public partial class MessageDisplay : Panel
{
    private VBoxContainer _messageContainer = null!;
    private PackedScene _errorMessageScene = null!;

    public override void _Ready()
    {
        _errorMessageScene = ResourceLoader.Load<PackedScene>(
            "res://scenes/components/error_message.tscn"
        );

        var messageContainerWrapper = ResourceLoader
            .Load<PackedScene>("res://scenes/components/message_container.tscn")
            .Instantiate();

        _messageContainer = messageContainerWrapper.GetChild<VBoxContainer>(0);

        GetTree().Root.CallDeferred("add_child", messageContainerWrapper);
    }

    public void ShowError(string errorMessage, params string[] subErrors)
    {
        var errorMessageNode = _errorMessageScene.Instantiate();

        var label =
            errorMessageNode.GetNode<Label>("MarginContainer/VBoxContainer/Label")
            ?? throw new Exception("Failed to get Label for error message node");

        var stringBuilder = new StringBuilder();
        stringBuilder.Append("Error: ");
        stringBuilder.Append(errorMessage);
        foreach (var error in subErrors)
        {
            stringBuilder.Append('\n');
            stringBuilder.Append('-');
            stringBuilder.Append(' ');
            stringBuilder.Append(error);
        }

        label.Text = stringBuilder.ToString();
        _messageContainer.AddChild(errorMessageNode);
    }
}
