using System;
using System.Text;
using Godot;
using Immerse.BfhClient.Utils;

namespace Immerse.BfhClient.UI;

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

        _messageContainer.CallDeferred(Strings.AddChild, errorMessageNode);
        button.Pressed += () =>
        {
            _messageContainer.CallDeferred(Strings.RemoveChild, errorMessageNode);
        };
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
