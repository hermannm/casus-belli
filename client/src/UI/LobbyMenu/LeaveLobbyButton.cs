using Godot;
using Immerse.BfhClient.Lobby;

namespace Immerse.BfhClient.UI.LobbyMenu;

public partial class LeaveLobbyButton : Button
{
    public override async void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
        await LobbyState.Instance.LeaveLobby();
    }
}
