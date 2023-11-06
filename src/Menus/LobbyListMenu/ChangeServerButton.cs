using Godot;

namespace Immerse.BfhClient.Menus.LobbyListMenu;

public partial class ChangeServerButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(Scenes.ServerAddressMenu);
    }
}
