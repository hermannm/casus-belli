using Godot;
using Immerse.BfhClient.Lobby;

namespace Immerse.BfhClient.UI.LobbyMenu;

public partial class StartGameButton : Button
{
    public override void _Ready()
    {
        UpdateButtonState();
        LobbyState.Instance.LobbyChangedSignal.Connect(UpdateButtonState);
    }

    private void UpdateButtonState()
    {
        if (LobbyState.Instance.ReadyToStartGame())
        {
            Disabled = false;
        }
        else
        {
            Disabled = true;
        }
    }
}
