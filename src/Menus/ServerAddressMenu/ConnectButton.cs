using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.ServerAddressMenu;

public partial class ConnectButton : Button
{
    private ApiClient _apiClient = null!;
    private TextEdit _serverAddressField = null!;

    public override void _Ready()
    {
        _apiClient = this.GetApiClient();
        _serverAddressField = GetNode<TextEdit>("%ServerAddressField");
    }

    public override void _Pressed()
    {
        if (!_apiClient.Connect(_serverAddressField.Text))
        {
            return;
        }

        var err = GetTree().ChangeSceneToFile(Scenes.LobbyListMenu);
        if (err != Error.Ok)
        {
            this.GetMessageDisplay().ShowError("Failed to load lobby list menu", err.ToString());
        }
    }
}
