using CasusBelli.Client.Lobby;
using Godot;

namespace CasusBelli.Client.UI.LobbyMenu;

public partial class StartGameButton : Button
{
    public override void _Ready()
    {
        UpdateButtonState();
        LobbyState.Instance.LobbyChanged += UpdateButtonState;
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
