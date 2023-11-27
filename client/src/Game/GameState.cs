using System;
using System.Collections.Generic;
using CasusBelli.Client.Api;
using CasusBelli.Client.Lobby;
using Godot;

namespace CasusBelli.Client.Game;

public partial class GameState : Node
{
    public static GameState Instance { get; private set; } = null!;

    public Season Season { get; private set; } = Season.Winter;
    public Phase CurrentPhase { get; private set; } = Phase.SubmittingOrders;
    public Dictionary<string, List<Order>> OrdersByFaction { get; private set; } = new();
    public List<Player> PlayersYetToSubmitOrders = new();
    public List<Battle> Battles = new();
    public List<DangerZoneCrossing> DangerZoneCrossings = new();

    [Signal]
    public delegate void PhaseChangedEventHandler();

    [Signal]
    public delegate void SupportCutEventHandler(string regionName);

    [Signal]
    public delegate void UncontestedMoveEventHandler(string fromRegion, string toRegion);

    private readonly Board _board = new();

    public enum Phase
    {
        SubmittingOrders,
        OrdersSubmitted,
        ResolvingOrders
    }

    public override void _EnterTree()
    {
        Instance = this;
        ApiClient.Instance.AddMessageHandler<GameStartedMessage>(HandleGameStarted);
        ApiClient.Instance.AddMessageHandler<OrderRequestMessage>(HandleOrderRequest);
        ApiClient.Instance.AddMessageHandler<OrdersConfirmationMessage>(HandleOrdersConfirmation);
        ApiClient.Instance.AddMessageHandler<OrdersReceivedMessage>(HandleOrdersReceived);
        ApiClient.Instance.AddMessageHandler<BattleResultsMessage>(HandleBattleResults);
        ApiClient.Instance.AddMessageHandler<DangerZoneCrossingsMessage>(HandleDangerZoneCrossings);

        LobbyState.Instance.LobbyChanged += () =>
        {
            PlayersYetToSubmitOrders = new List<Player>(LobbyState.Instance.OtherPlayers);
        };
    }

    public override void _Process(double delta) { }

    private void HandleGameStarted(GameStartedMessage message)
    {
        _board.Regions = message.Board;
    }

    private void HandleOrderRequest(OrderRequestMessage message)
    {
        Season = message.Season;
        CurrentPhase = Phase.SubmittingOrders;
        EmitSignal(SignalName.PhaseChanged);
        OrdersByFaction.Clear();
        PlayersYetToSubmitOrders = new List<Player>(LobbyState.Instance.OtherPlayers);
    }

    private void HandleOrdersConfirmation(OrdersConfirmationMessage message)
    {
        if (message.FactionThatSubmittedOrders == LobbyState.Instance.Player.Faction)
        {
            CurrentPhase = Phase.OrdersSubmitted;
            EmitSignal(SignalName.PhaseChanged);
        }
        else
        {
            PlayersYetToSubmitOrders.RemoveAll(
                player => player.Faction == message.FactionThatSubmittedOrders
            );
        }
    }

    private void HandleOrdersReceived(OrdersReceivedMessage message)
    {
        OrdersByFaction = message.OrdersByFaction;
        PlaceOrdersOnBoard();

        CurrentPhase = Phase.ResolvingOrders;
        EmitSignal(SignalName.PhaseChanged);

        ResolveMoves();
    }

    private void HandleBattleResults(BattleResultsMessage message)
    {
        throw new NotImplementedException();
    }

    private void HandleDangerZoneCrossings(DangerZoneCrossingsMessage message)
    {
        DangerZoneCrossings.AddRange(message.Crossings);
    }

    public void PlaceOrdersOnBoard()
    {
        var supportOrders = new List<Order>();

        foreach (var (_, factionOrders) in OrdersByFaction)
        {
            foreach (var order in factionOrders)
            {
                if (order.Type == OrderType.Support)
                {
                    supportOrders.Add(order);
                    continue;
                }

                _board.PlaceOrder(order);
            }
        }

        foreach (var supportOrder in supportOrders)
        {
            if (!_board.Regions[supportOrder.Origin].Attacked())
            {
                _board.PlaceOrder(supportOrder);
            }
            else
            {
                EmitSignal(SignalName.SupportCut, supportOrder.Origin);
            }
        }
    }

    private void ResolveMoves()
    {
        var allRegionsWaiting = false;
        while (!allRegionsWaiting)
        {
            allRegionsWaiting = true;

            foreach (var (_, region) in _board!.Regions)
            {
                var waiting = ResolveRegionMoves(region);
                if (!waiting)
                {
                    allRegionsWaiting = false;
                }
            }
        }
    }

    private bool ResolveRegionMoves(Region region)
    {
        return false;
    }
}
