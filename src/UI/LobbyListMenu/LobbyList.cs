using System.Threading.Tasks;
using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.Lobby;

namespace Immerse.BfhClient.UI.LobbyListMenu;

public partial class LobbyList : Node
{
    private PackedScene _lobbyListItemScene = ResourceLoader.Load<PackedScene>(
        Scenes.LobbyListItem
    );
    private Popup _usernameInputPopup = null!;
    private LobbyInfo? _lobbyToJoin = null;

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
        usernameInput.Text = LobbyState.Instance.Player.Username;

        var joinLobbyButton = _usernameInputPopup.GetNode<Button>("%JoinLobbyButton");
        joinLobbyButton.Pressed += async () =>
        {
            if (_lobbyToJoin is { } lobby)
            {
                var success = await LobbyState.Instance.TryJoinLobby(lobby, usernameInput.Text);
                if (success)
                {
                    SceneManager.Instance.LoadScene(Scenes.LobbyMenu);
                }
            }
            else
            {
                GD.PushError("Lobby to join is not set");
            }
        };
    }
}
