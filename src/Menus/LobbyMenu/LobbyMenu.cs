using System.Linq;
using Godot;
using Immerse.BfhClient.Lobby;
using Immerse.BfhClient.Utils;

namespace Immerse.BfhClient.Menus.LobbyMenu;

public partial class LobbyMenu : Node
{
    private OptionButton _factionSelect = null!; // Set in _Ready
    private Node _otherPlayersList = null!; // Set in _Ready
    private Button _startGameButton = null!; // Set in _Ready
    private PackedScene _playerListItemScene = ResourceLoader.Load<PackedScene>(
        Scenes.PlayerListItem
    );

    public override void _Ready()
    {
        InitializeActivePlayer();
        _otherPlayersList = GetNode("%OtherPlayerList");
        _startGameButton = GetNode<Button>("%StartGameButton");
        GetNode<Button>("%LeaveLobbyButton").Pressed += LeaveLobby;
        UpdateLobbyView();
        LobbyState.Instance.ConnectSignal(LobbyState.LobbyChangeSignal, UpdateLobbyView);
    }

    private void InitializeActivePlayer()
    {
        var activePlayerItem = GetNode("%ActivePlayer");
        activePlayerItem.GetNode<Label>("%Username").Text = LobbyState.Instance.Player.Username;

        _factionSelect = new OptionButton();
        _factionSelect.AddItem("None selected", 0);
        _factionSelect.Select(0);

        _factionSelect.ItemSelected += SelectFaction;
        activePlayerItem.GetNode("%FactionContainer").AddChild(_factionSelect);
    }

    private void UpdateLobbyView()
    {
        _otherPlayersList.ClearChildren();

        UpdateActivePlayerSelectedFaction();
        foreach (var player in LobbyState.Instance.OtherPlayers)
        {
            AddPlayerListItem(player);
        }

        if (LobbyState.Instance.ReadyToStartGame())
        {
            _startGameButton.Disabled = false;
        }
        else
        {
            _startGameButton.Disabled = true;
        }
    }

    private void UpdateActivePlayerSelectedFaction()
    {
        if (_factionSelect.ItemCount == 1)
        {
            var index = 1;
            foreach (var faction in LobbyState.Instance.SelectableFactions)
            {
                _factionSelect.AddItem(faction, index);
                index++;
            }
        }

        var selectedFactionIndex = 0;
        if (LobbyState.Instance.Player.Faction != null)
        {
            selectedFactionIndex =
                LobbyState.Instance.SelectableFactions.IndexOf(LobbyState.Instance.Player.Faction)
                + 1; // +1 for "None selected" first option
        }

        if (_factionSelect.Selected != selectedFactionIndex)
        {
            // Temporarily unsubscribes signal handler to avoid re-sending SelectFaction message
            _factionSelect.ItemSelected -= SelectFaction;
            _factionSelect.Select(selectedFactionIndex);
            _factionSelect.ItemSelected += SelectFaction;
        }
    }

    private void AddPlayerListItem(Player player)
    {
        var listItem = _playerListItemScene.Instantiate();
        listItem.GetNode<Label>("%Username").Text = player.Username;

        var faction = new Label();
        faction.Text = player.Faction ?? "None selected";
        listItem.GetNode("%FactionContainer").AddChild(faction);
    }

    private static void SelectFaction(long selectedFactionIndex)
    {
        var selectedFaction = LobbyState.Instance.SelectableFactions.ElementAtOrDefault(
            (int)selectedFactionIndex - 1 // -1 to account for "None selected" first option
        );
        LobbyState.SelectFaction(selectedFaction);
    }

    private static async void LeaveLobby()
    {
        SceneManager.Instance.LoadPreviousScene();
        await LobbyState.Instance.LeaveLobby();
    }
}
