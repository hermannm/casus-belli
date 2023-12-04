using Godot;

namespace CasusBelli.Client.Menus.MainMenu;

public partial class QuitButton : Button
{
    public override void _Pressed()
    {
        GetTree().Quit();
    }
}
