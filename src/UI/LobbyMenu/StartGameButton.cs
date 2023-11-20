using Godot;
using Immerse.BfhClient.Lobby;
using Immerse.BfhClient.Utils;

namespace Immerse.BfhClient.UI.LobbyMenu;

public partial class StartGameButton : Button
{
    public override void _Ready()
    {
        UpdateButtonState();
        LobbyState.Instance.ConnectSignal(LobbyState.LobbyChangeSignal, UpdateButtonState);
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
