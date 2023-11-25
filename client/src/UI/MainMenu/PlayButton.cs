using CasusBelli.Client.Api;
using Godot;

namespace CasusBelli.Client.UI.MainMenu;

public partial class PlayButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(
            ApiClient.Instance.ServerUrl is null ? Scenes.ServerAddressMenu : Scenes.LobbyListMenu
        );
    }
}
