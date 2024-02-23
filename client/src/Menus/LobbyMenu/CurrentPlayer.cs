using System.Linq;
using CasusBelli.Client.Lobby;
using CasusBelli.Client.Utils;
using Godot;

namespace CasusBelli.Client.Menus.LobbyMenu;

public partial class CurrentPlayer : Node
{
    private OptionButton _factionSelect = null!; // Set in _Ready

    public override void _Ready()
    {
        GetNode<Label>("%Username").Text = LobbyState.Instance.Player.Username;

        _factionSelect = new OptionButton();
        _factionSelect.AddItem("None selected", 0);
        _factionSelect.Select(0);
        _factionSelect.ItemSelected += SelectFaction;
        _factionSelect.AddThemeFontSizeOverride(Strings.FontSize, 20);
        GetNode("%FactionContainer").AddChild(_factionSelect);

        // Uses Connect() instead of += syntax, because of this issue:
        // https://github.com/godotengine/godot/issues/70414
        LobbyState.Instance.Connect(
            LobbyState.SignalName.LobbyChanged,
            Callable.From(UpdateSelectedFaction)
        );
    }

    private static void SelectFaction(long selectedFactionIndex)
    {
        var selectedFaction = LobbyState.Instance.SelectableFactions.ElementAtOrDefault(
            (int)selectedFactionIndex - 1 // -1 to account for "None selected" first option
        );
        LobbyState.SelectFaction(selectedFaction);
    }

    private void UpdateSelectedFaction()
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
}
