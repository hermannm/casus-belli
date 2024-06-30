using CasusBelli.Client.Lobby;
using Godot;

namespace CasusBelli.Client.UI.LobbyMenu;

public partial class LeaveLobbyButton : Button
{
    public override async void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
        await LobbyState.Instance.LeaveLobby();
    }
}
