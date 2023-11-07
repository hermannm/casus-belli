using Godot;
using Immerse.BfhClient.Api;

namespace Immerse.BfhClient.Menus.ServerAddressMenu;

public partial class ConnectButton : Button
{
    private TextEdit _serverAddressField = null!; // Set in _Ready

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

        SceneManager.Instance.LoadScene(Scenes.LobbyListMenu);
    }
}
