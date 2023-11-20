using Godot;

namespace Immerse.BfhClient.UI.LobbyListMenu;

public partial class MainMenuButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(Scenes.MainMenu);
    }
}
