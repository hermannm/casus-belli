using System;
using System.Text;
using CasusBelli.Client.Utils;
using Godot;

namespace CasusBelli.Client.UI;

public partial class MessageDisplay : Node
{
    /// MessageDisplay singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static MessageDisplay Instance { get; private set; } = null!;

    private ScrollContainer _scrollContainer = null!; // Set in _Ready
    private VBoxContainer _messageContainer = null!; // Set in _Ready
    private const int MaxScrollContainerHeight = 1040; // 1080 - 2x20 margins

    public override void _EnterTree()
    {
        Instance = this;
    }

    public override void _Ready()
    {
        _scrollContainer = GetNode<ScrollContainer>("%ScrollContainer");
        _messageContainer = GetNode<VBoxContainer>("%MessageContainer");
        _messageContainer.Resized += ResizeScrollContainer;
    }

    public void ShowError(string errorMessage, params string[] subErrors)
    {
        var message = ShowMessage(ScenePaths.ErrorMessage, "Error: ", errorMessage, subErrors);
        GD.Print(message);
    }

    public void ShowInfo(string infoMessage, params string[] subMessages)
    {
        ShowMessage(ScenePaths.InfoMessage, null, infoMessage, subMessages);
    }

    private string ShowMessage(string scene, string? prefix, string message, string[] subMessages)
    {
        var errorMessageNode = ResourceLoader.Load<PackedScene>(scene).Instantiate();
        var label = errorMessageNode.GetNode<Label>("%MessageLabel");
        var closeButton = errorMessageNode.GetNode<TextureButton>("%CloseButton");

        var stringBuilder = new StringBuilder();
        if (prefix is not null)
        {
            stringBuilder.Append(prefix);
        }
        if (message.Length >= 2)
        {
            stringBuilder.Append(char.ToUpper(message[0]));
            stringBuilder.Append(message.AsSpan(1));
        }
        foreach (var subMessage in subMessages)
        {
            stringBuilder.Append('\n');
            stringBuilder.Append('-');
            stringBuilder.Append(' ');
            stringBuilder.Append(subMessage);
        }

        message = stringBuilder.ToString();
        label.Text = message;

        _messageContainer.CallDeferred(Strings.AddChild, errorMessageNode);
        closeButton.Pressed += () =>
        {
            errorMessageNode.QueueFree();
        };

        return message;
    }

    private void ResizeScrollContainer()
    {
        var newHeight = _messageContainer.Size.Y;
        if (newHeight > MaxScrollContainerHeight)
        {
            newHeight = MaxScrollContainerHeight;
        }
        _scrollContainer.Size = new Vector2(_scrollContainer.Size.X, newHeight);
    }
}
