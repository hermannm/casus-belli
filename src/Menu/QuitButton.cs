using Godot;

namespace Immerse.BfhClient.Menu;

public partial class QuitButton : Button
{
    public override void _Pressed()
    {
        GetTree().Quit();
    }
}
