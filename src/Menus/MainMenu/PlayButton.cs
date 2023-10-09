using Godot;

namespace Immerse.BfhClient.Menus.MainMenu;

public partial class PlayButton : Button
{
    public override void _Pressed()
    {
        GetTree().ChangeSceneToFile(Scenes.ServerAddressMenu);
    }
}
