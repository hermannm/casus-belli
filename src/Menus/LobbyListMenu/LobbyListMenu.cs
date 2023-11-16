using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.LobbyListMenu;

public partial class LobbyListMenu : Node
{
    public override void _EnterTree()
    {
        ApiClient.Instance.AddMessageHandler<LobbyJoinedMessage>(HandleLobbyJoinedMessage);
    }

    public override async void _Ready()
    {
        await ApiClient.Instance.TryJoinLobby(
            new LobbyInfo
            {
                Name = "test",
                BoardInfo = new BoardInfo
                {
                    Id = "bfh_5players",
                    Name = "The Battle for Hermannia (5 players)",
                    WinningCastleCount = 5
                }
            },
            "hermannm"
        );
    }

    private static void HandleLobbyJoinedMessage(LobbyJoinedMessage message)
    {
        MessageDisplay.Instance.ShowError("Received lobby joined message!");
    }
}
