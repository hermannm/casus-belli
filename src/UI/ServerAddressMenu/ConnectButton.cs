using Godot;
using Immerse.BfhClient.Api;

namespace Immerse.BfhClient.UI.ServerAddressMenu;

public partial class ConnectButton : Button
{
    private LineEdit _serverAddressField = null!; // Set in _Ready

    public override void _Ready()
    {
        _serverAddressField = GetNode<LineEdit>("%ServerAddressInput");
    }

    public override void _Pressed()
    {
        if (!ApiClient.Instance.TryConnect(_serverAddressField.Text))
        {
            return;
        }

        SceneManager.Instance.LoadScene(Scenes.LobbyListMenu);
    }
}
