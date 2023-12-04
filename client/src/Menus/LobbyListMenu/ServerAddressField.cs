using CasusBelli.Client.Api;
using Godot;

namespace CasusBelli.Client.Menus.LobbyListMenu;

public partial class ServerAddressField : Label
{
    public override void _Ready()
    {
        var serverUrl = ApiClient.Instance.ServerUrl;
        if (serverUrl is null)
        {
            MessageDisplay.Instance.ShowError("Failed to show server URL", "URL was null");
            return;
        }

        Text += serverUrl.ToString();
    }
}
