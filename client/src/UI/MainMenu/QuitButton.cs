using Godot;

namespace Immerse.BfhClient.UI.MainMenu;

public partial class QuitButton : Button
{
    public override void _Pressed()
    {
        GetTree().Quit();
    }
}
