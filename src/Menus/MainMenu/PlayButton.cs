using Godot;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.MainMenu;

public partial class PlayButton : Button
{
    public override void _Pressed()
    {
        var err = GetTree().ChangeSceneToFile(Scenes.ServerAddressMenu);
        if (err != Error.Ok)
        {
            this.GetMessageDisplay()
                .ShowError("Failed to load server address menu", err.ToString());
        }
    }
}
