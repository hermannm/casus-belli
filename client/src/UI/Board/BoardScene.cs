using System;
using CasusBelli.Client.Lobby;
using Godot;

namespace CasusBelli.Client.UI.Board;

public partial class BoardScene : Node
{
    private ColorMap _colorMap = null!;

    public override void _Ready()
    {
        var boardId = LobbyState.Instance.BoardId;
        if (boardId is null)
        {
            throw new Exception("Loaded board scene while lobby board ID was still null");
        }

        _colorMap = new ColorMap(boardId);
    }
}
