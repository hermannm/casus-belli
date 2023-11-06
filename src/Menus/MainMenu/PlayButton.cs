using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.MainMenu;

public partial class PlayButton : Button
{
    public override void _Pressed()
    {
        if (ApiClient.Instance.ServerUrl is null)
        {
            var err = GetTree().ChangeSceneToFile(Scenes.ServerAddressMenu);
            if (err != Error.Ok)
            {
                MessageDisplay.Instance.ShowError(
                    "Failed to load server address menu",
                    err.ToString()
                );
            }
        }
        else
        {
            var err = GetTree().ChangeSceneToFile(Scenes.LobbyListMenu);
            if (err != Error.Ok)
            {
                MessageDisplay.Instance.ShowError("Failed to load lobby list menu", err.ToString());
            }
        }
    }
}
