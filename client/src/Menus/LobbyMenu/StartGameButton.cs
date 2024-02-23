using CasusBelli.Client.Game;
using CasusBelli.Client.Lobby;
using Godot;

namespace CasusBelli.Client.Menus.LobbyMenu;

public partial class StartGameButton : Button
{
    public override void _Pressed()
    {
        GameState.StartGame();
    }

    public override void _Ready()
    {
        UpdateButtonState();
        LobbyState.Instance.Connect(
            LobbyState.SignalName.LobbyChanged,
            Callable.From(UpdateButtonState)
        );
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
