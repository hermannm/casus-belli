using Godot;

namespace CasusBelli.Client.UI.LobbyListMenu;

public partial class ChangeServerButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadScene(ScenePaths.ServerAddressMenu);
    }
}
