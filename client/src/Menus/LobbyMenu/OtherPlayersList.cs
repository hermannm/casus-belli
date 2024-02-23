using CasusBelli.Client.Lobby;
using CasusBelli.Client.Utils;
using Godot;

namespace CasusBelli.Client.Menus.LobbyMenu;

public partial class OtherPlayersList : Node
{
    private PackedScene _playerListItemScene = ResourceLoader.Load<PackedScene>(
        Scenes.PlayerListItem
    );

    public override void _Ready()
    {
        UpdatePlayerList();
        LobbyState.Instance.Connect(
            LobbyState.SignalName.LobbyChanged,
            Callable.From(UpdatePlayerList)
        );
    }

    private void UpdatePlayerList()
    {
        this.ClearChildren();

        foreach (var player in LobbyState.Instance.OtherPlayers)
        {
            var listItem = _playerListItemScene.Instantiate();
            listItem.GetNode<Label>("%Username").Text = player.Username;

            var faction = new Label();
            faction.Text = player.Faction ?? "None selected";
            faction.AddThemeFontSizeOverride(Strings.FontSize, 20);
            listItem.GetNode("%FactionContainer").AddChild(faction);

            AddChild(listItem);
        }
    }
}
