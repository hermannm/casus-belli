using Godot;
using Immerse.BfhClient.Api;

namespace Immerse.BfhClient.UI.LobbyListMenu;

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
