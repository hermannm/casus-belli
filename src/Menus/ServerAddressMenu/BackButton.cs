using Godot;

namespace Immerse.BfhClient.Menus.ServerAddressMenu;

public partial class BackButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
    }
}
