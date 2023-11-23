using System.Collections.Generic;
using System.Threading.Tasks;
using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.Utils;

namespace Immerse.BfhClient.Lobby;

public partial class LobbyState : Node
{
    public static LobbyState Instance { get; private set; } = null!;

    public Player Player { get; private set; } = new();
    public List<Player> OtherPlayers { get; private set; } = new();
    public List<string> SelectableFactions { get; private set; } = new();
    public CustomSignal LobbyChangedSignal { get; } = new("LobbyChanged");

    private LobbyInfo? _joinedLobby = null;

    public override void _EnterTree()
    {
        Instance = this;
        ApiClient.Instance.AddMessageHandler<LobbyJoinedMessage>(HandleLobbyJoinedMessage);
        ApiClient.Instance.AddMessageHandler<PlayerStatusMessage>(HandlePlayerStatusMessage);
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

    private void HandleLobbyJoinedMessage(LobbyJoinedMessage message)
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

        LobbyChangedSignal.Emit();
    }

    private void HandlePlayerStatusMessage(PlayerStatusMessage message)
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

        LobbyChangedSignal.Emit();
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
