using System.Collections.Generic;
using System.Threading.Tasks;
using CasusBelli.Client.Api;
using CasusBelli.Client.Api.Messages;
using Godot;

namespace CasusBelli.Client.Lobby;

public partial class LobbyState : Node
{
    public static LobbyState Instance { get; private set; } = null!;

    public Player Player { get; private set; } = new();
    public List<Player> OtherPlayers { get; private set; } = new();
    public List<string> SelectableFactions { get; private set; } = new();

    [Signal]
    public delegate void LobbyChangedEventHandler();

    private LobbyInfo? _joinedLobby = null;

    public override void _EnterTree()
    {
        Instance = this;
        ApiClient.Instance.AddMessageHandler<LobbyJoinedMessage>(HandleLobbyJoined);
        ApiClient.Instance.AddMessageHandler<PlayerStatusMessage>(HandlePlayerStatus);
    }

    public async Task<bool> TryJoinLobby(LobbyInfo lobby, string username)
    {
        Player.Username = username;
        var success = await ApiClient.Instance.TryJoinLobby(lobby.Name, username);
        if (success)
        {
            _joinedLobby = lobby;
        }

        return success;
    }

    public static void SelectFaction(string? faction)
    {
        ApiClient.Instance.SendMessage(new SelectFactionMessage { Faction = faction });
    }

    public bool ReadyToStartGame()
    {
        if (
            Player.Faction is null
            || _joinedLobby is null
            || OtherPlayers.Count + 1 < _joinedLobby.BoardInfo.PlayerFactions.Count
        )
        {
            return false;
        }

        foreach (var player in OtherPlayers)
        {
            if (player.Faction is null)
            {
                return false;
            }
        }

        return true;
    }

    public Task LeaveLobby()
    {
        _joinedLobby = null;
        return ApiClient.Instance.LeaveLobby();
    }

    private void HandleLobbyJoined(LobbyJoinedMessage message)
    {
        SelectableFactions = message.SelectableFactions;
        foreach (var playerStatus in message.PlayerStatuses)
        {
            OtherPlayers.Add(
                new Player
                {
                    Username = playerStatus.Username,
                    Faction = playerStatus.SelectedFaction
                }
            );
        }

        EmitSignal(SignalName.LobbyChanged);
    }

    private void HandlePlayerStatus(PlayerStatusMessage message)
    {
        if (GetPlayerByUsername(message.Username) is { } player)
        {
            player.Faction = message.SelectedFaction;
        }
        else
        {
            OtherPlayers.Add(
                new Player { Username = message.Username, Faction = message.SelectedFaction }
            );
        }

        EmitSignal(SignalName.LobbyChanged);
    }

    private Player? GetPlayerByUsername(string username)
    {
        if (username == Player.Username)
        {
            return Player;
        }

        foreach (var player in OtherPlayers)
        {
            if (player.Username == username)
            {
                return player;
            }
        }

        return null;
    }
}
