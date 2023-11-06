using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.MainMenu;

public partial class PlayButton : Button
{
    public override void _Pressed()
    {
        MessageDisplay.Instance.ShowError("test");
        if (ApiClient.Instance.ServerUrl is null)
        {
            SceneManager.Instance.LoadScene(Scenes.ServerAddressMenu);
        }
        else
        {
            SceneManager.Instance.LoadScene(Scenes.LobbyListMenu);
        }
    }
}
