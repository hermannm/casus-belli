using Godot;

namespace CasusBelli.Client.UI.MainMenu;

public partial class QuitButton : Button
{
    public override void _Pressed()
    {
        GetTree().Quit();
    }
}
