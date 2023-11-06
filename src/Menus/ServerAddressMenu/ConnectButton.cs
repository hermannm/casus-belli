using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.ServerAddressMenu;

public partial class ConnectButton : Button
{
    private TextEdit _serverAddressField = null!;

    public override void _Ready()
    {
        _serverAddressField = GetNode<TextEdit>("%ServerAddressField");
    }

    public override void _Pressed()
    {
        if (!ApiClient.Instance.Connect(_serverAddressField.Text))
        {
            return;
        }

        var err = GetTree().ChangeSceneToFile(Scenes.LobbyListMenu);
        if (err != Error.Ok)
        {
            MessageDisplay.Instance.ShowError("Failed to load lobby list menu", err.ToString());
        }
    }
}
