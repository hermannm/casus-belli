using Godot;

namespace Immerse.BfhClient.UI.LobbyListMenu;

public partial class ChangeServerButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(Scenes.ServerAddressMenu);
    }
}
