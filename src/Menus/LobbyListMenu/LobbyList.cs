using System.Threading.Tasks;
using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient.Menus.LobbyListMenu;

public partial class LobbyList : Node
{
    private PackedScene _lobbyListItemScene = ResourceLoader.Load<PackedScene>(
        Scenes.LobbyListItem
    );
    private Popup _usernameInputPopup = null!;
    private LobbyInfo? _lobbyToJoin = null;

    public override void _EnterTree()
    {
        ApiClient.Instance.AddMessageHandler<LobbyJoinedMessage>(HandleLobbyJoinedMessage);
    }

    public override async void _Ready()
    {
        _usernameInputPopup = GetNode<Popup>("%UsernameInputPopup");
        await PopulateLobbyList();
        PrepareUsernameInputPopup();
    }

    private async Task PopulateLobbyList()
    {
        var lobbies = await ApiClient.Instance.ListLobbies();
        if (lobbies is null)
            return;

        foreach (var lobby in lobbies)
        {
            var lobbyListItem = _lobbyListItemScene.Instantiate();

            lobbyListItem.GetNode<Label>("%LobbyName").Text = lobby.Name;
            lobbyListItem.GetNode<Label>("%PlayerCount").Text =
                $"{lobby.PlayerCount}/{lobby.BoardInfo.PlayerFactions.Count} players";

            var joinButton = lobbyListItem.GetNode<Button>("%JoinButton");
            if (lobby.PlayerCount < lobby.BoardInfo.PlayerFactions.Count)
            {
                joinButton.Pressed += () =>
                {
                    _lobbyToJoin = lobby;
                    _usernameInputPopup.PopupCentered();
                };
            }
            else
            {
                joinButton.Disabled = true;
            }

            AddChild(lobbyListItem);
        }
    }

    private void PrepareUsernameInputPopup()
    {
        var usernameInput = _usernameInputPopup.GetNode<LineEdit>("%UsernameInput");
        var joinLobbyButton = _usernameInputPopup.GetNode<Button>("%JoinLobbyButton");

        joinLobbyButton.Pressed += async () =>
        {
            if (_lobbyToJoin is { } lobby)
            {
                var username = usernameInput.Text;
                await ApiClient.Instance.TryJoinLobby(lobby, username);
            }
            else
            {
                GD.PushError("Lobby to join is not set");
            }
        };
    }

    private static void HandleLobbyJoinedMessage(LobbyJoinedMessage message)
    {
        MessageDisplay.Instance.ShowInfo("Received lobby joined message!");
    }
}
