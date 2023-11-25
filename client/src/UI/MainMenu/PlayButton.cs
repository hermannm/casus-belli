using Godot;
using Immerse.BfhClient.Api;

namespace Immerse.BfhClient.UI.MainMenu;

public partial class PlayButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(
            ApiClient.Instance.ServerUrl is null ? Scenes.ServerAddressMenu : Scenes.LobbyListMenu
        );
    }
}
