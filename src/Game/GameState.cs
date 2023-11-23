using System.Collections.Generic;
using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.Lobby;
using Immerse.BfhClient.UI;
using Immerse.BfhClient.Utils;

namespace Immerse.BfhClient.Game;

public partial class GameState : Node
{
    public static GameState Instance { get; private set; }

    public Season Season { get; private set; } = Season.Winter;
    public GamePhase Phase { get; private set; } = GamePhase.SubmittingOrders;
    public Dictionary<string, List<Order>> OrdersByFaction { get; private set; } = new();
    public List<Player> PlayersYetToSubmitOrders = new();
    public List<Battle> Battles = new();

    public CustomSignal PhaseChangedSignal = new("PhaseChanged");
    public CustomSignal<SupportCut> SupportCutSignal = new("SupportCut");
    public CustomSignal<UncontestedMove> UncontestedMoveSignal = new("UncontestedMove");

    private Board? _board = null;
    private List<Battle> _unprocessedBattles = new();

    public override void _EnterTree()
    {
        Instance = this;
        ApiClient.Instance.AddMessageHandler<GameStartedMessage>(HandleGameStartedMessage);
        ApiClient.Instance.AddMessageHandler<OrderRequestMessage>(HandleOrderRequestMessage);
        ApiClient.Instance.AddMessageHandler<OrdersConfirmationMessage>(
            HandleOrdersConfirmationMessage
        );
        ApiClient.Instance.AddMessageHandler<OrdersReceivedMessage>(HandleOrdersReceivedMessage);
        ApiClient.Instance.AddMessageHandler<BattleResultsMessage>(HandleBattleResultsMessage);

        LobbyState.Instance.LobbyChangedSignal.Connect(() =>
        {
            PlayersYetToSubmitOrders = new List<Player>(LobbyState.Instance.OtherPlayers);
        });
    }

    public override void _Process(double delta) { }

    private void HandleGameStartedMessage(GameStartedMessage message)
    {
        _board = message.Board;
    }

    private void HandleOrderRequestMessage(OrderRequestMessage message)
    {
        Season = message.Season;
        Phase = GamePhase.SubmittingOrders;
        PhaseChangedSignal.Emit();
        OrdersByFaction.Clear();
        PlayersYetToSubmitOrders = new List<Player>(LobbyState.Instance.OtherPlayers);
    }

    private void HandleOrdersConfirmationMessage(OrdersConfirmationMessage message)
    {
        if (message.FactionThatSubmittedOrders == LobbyState.Instance.Player.Faction)
        {
            Phase = GamePhase.OrdersSubmitted;
            PhaseChangedSignal.Emit();
        }
        else
        {
            PlayersYetToSubmitOrders.RemoveAll(
                player => player.Faction == message.FactionThatSubmittedOrders
            );
        }
    }

    private void HandleOrdersReceivedMessage(OrdersReceivedMessage message)
    {
        if (_board is null)
        {
            MessageDisplay.Instance.ShowError(
                "Tried to resolve moves before board was initialized"
            );
            return;
        }

        OrdersByFaction = message.OrdersByFaction;
        PlaceOrdersOnBoard();

        Phase = GamePhase.ResolvingOrders;
        PhaseChangedSignal.Emit();

        ResolveUncontestedMoves();
    }

    private void HandleBattleResultsMessage(BattleResultsMessage message)
    {
        if (_board is null)
        {
            MessageDisplay.Instance.ShowError(
                "Tried to resolve moves before board was initialized"
            );
            return;
        }

        _unprocessedBattles.AddRange(message.Battles);
    }

    private void PlaceOrdersOnBoard()
    {
        foreach (var (_, factionOrders) in OrdersByFaction)
        {
            foreach (var order in factionOrders)
            {
                var origin = _board![order.Origin];
                if (origin.Unit is { } unit)
                {
                    order.UnitType = unit.Type;
                    origin.Order = order;

                    if (order is { Type: OrderType.Move, Destination: not null })
                    {
                        _board[order.Destination].IncomingMoves.Add(order);
                    }
                }
                else
                {
                    MessageDisplay.Instance.ShowError(
                        "Received order for board region without unit"
                    );
                }
            }
        }

        foreach (var (_, region) in _board!)
        {
            if (region.Order?.Type == OrderType.Support && region.Attacked())
            {
                region.Order = null;
                SupportCutSignal.Emit(new SupportCut { RegionName = region.Name });
            }
        }
    }

    private void ResolveUncontestedMoves()
    {
        var allRegionsWaiting = false;
        while (allRegionsWaiting)
        {
            allRegionsWaiting = true;

            foreach (var (_, region) in _board!)
            {
                if (region.Empty() && region.Controlled() && region.IncomingMoves.Count == 1)
                {
                    var move = region.IncomingMoves[0];
                    _board[move.Origin].MoveUnitTo(region);

                    UncontestedMoveSignal.Emit(
                        new UncontestedMove { FromRegion = move.Origin, ToRegion = region.Name }
                    );

                    allRegionsWaiting = true;
                }
            }
        }
    }

    private void SucceedMove(Order move)
    {
        var destination = _board![move.Destination!];

        destination.Unit = move.Unit();
        destination.Order = null;
        if (!destination.Sea)
        {
            destination.ControllingFaction = move.Faction;
        }

        _board[move.Origin].RemoveUnit(move.Unit());
        _board.RemoveOrder(move);

        destination.Resolved = true;
    }
}
