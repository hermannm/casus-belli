using System;
using System.Text;
using Godot;

namespace Immerse.BfhClient.UI;

public partial class MessageDisplay : Panel
{
    /// MessageDisplay singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static MessageDisplay Instance { get; private set; } = null!;

    private VBoxContainer _messageContainer = null!;
    private PackedScene _errorMessageScene = null!;

    public override void _EnterTree()
    {
        // ReSharper disable once ConditionIsAlwaysTrueOrFalseAccordingToNullableAPIContract
        if (Instance is null)
        {
            Instance = this;
        }
    }

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
        _messageContainer.AddChild(errorMessageNode);

        button.Pressed += () =>
        {
            _messageContainer.RemoveChild(errorMessageNode);
        };
    }
}
