using Godot;

namespace Immerse.BfhClient.Menus.MainMenu;

public partial class QuitButton : Button
{
    public override void _Pressed()
    {
        GetTree().Quit();
    }
}
