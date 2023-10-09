using Godot;

namespace Immerse.BfhClient.Menus.ServerAddressMenu;

public partial class BackButton : Button
{
    public override void _Pressed()
    {
        GetTree().ChangeSceneToFile(Scenes.MainMenu);
    }
}
