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
        _errorMessageScene = ResourceLoader.Load<PackedScene>(Scenes.ErrorMessage);

        var messageDisplay =
            ResourceLoader.Load<PackedScene>(Scenes.MessageDisplay).Instantiate()
            ?? throw new Exception("Failed to instantiate message display scene");

        _messageContainer =
            messageDisplay.GetNode<VBoxContainer>("%MessageContainer")
            ?? throw new Exception("Failed to get message container for message display");

        GetTree().Root.CallDeferred("add_child", messageDisplay);
    }

    public void ShowError(string errorMessage, params string[] subErrors)
    {
        var errorMessageNode = _errorMessageScene.Instantiate();

        var label =
            errorMessageNode.GetNode<Label>("%ErrorMessageLabel")
            ?? throw new Exception("Failed to get label for error message node");

        var button =
            errorMessageNode.GetNode<TextureButton>("%CloseErrorButton")
            ?? throw new Exception("Failed to get close button for error message node");

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

        button.Pressed += () =>
        {
            _messageContainer.RemoveChild(errorMessageNode);
        };
    }
}
