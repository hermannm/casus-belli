using Godot;

namespace CasusBelli.Client.Menus.LobbyListMenu;

public partial class MainMenuButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(Scenes.MainMenu);
    }
}
