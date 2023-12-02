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

        if (Season == Season.Winter)
        {
            ResolveWinterOrders();
        }
        else
        {
            ResolveUncontestedRegions();
        }
    }

    private void HandleBattleResults(BattleResultsMessage message)
    {
        Battles.Add(message.Battle);
        ResolveBattle(message.Battle);
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

    private void ResolveBattle(Battle battle)
    {
        var (isDangerZoneCrossing, succeeded, order) = battle.IsDangerZoneCrossing();
        if (isDangerZoneCrossing && !succeeded)
        {
            if (order!.Type == OrderType.Move)
            {
                _board.KillMove(order);
            }
            else
            {
                _board.RemoveOrder(order);
            }
            return;
        }
    }

    private void ResolveUncontestedRegions()
    {
        var allRegionsWaiting = false;
        while (!allRegionsWaiting)
        {
            allRegionsWaiting = true;

            foreach (var (_, region) in _board!.Regions)
            {
                var waiting = ResolveUncontestedRegion(region);
                if (!waiting)
                {
                    allRegionsWaiting = false;
                }
            }
        }
    }

    /// <returns>Whether the region is waiting for a battle to resolve further.</returns>
    private bool ResolveUncontestedRegion(Region region)
    {
        var mustWait = TransportResolver.ResolveUncontestedTransports(region, _board);
        if (mustWait)
        {
            return true;
        }

        if (!region.Attacked())
        {
            region.ResolveRetreat();

            if (region.ExpectedKnightMoves == 0)
            {
                region.Resolved = true;
                return true;
            }
            else if (region.ExpectedKnightMoves == region.IncomingKnightMoves.Count)
            {
                _board.PlaceKnightMoves(region);
                return false;
            }
        }

        if (_board.FindBorderBattle(region))
        {
            return true;
        }

        if (!region.PartOfCycle)
        {
            var cycle = _board.FindCycle(region.Name, region);
            if (cycle is not null)
            {
                Board.PrepareCycleForResolving(cycle);
                return false;
            }
        }

        if (
            region.IncomingMoves.Count == 1 && region.Empty() && (region.Controlled() || region.Sea)
        )
        {
            var move = region.IncomingMoves[0];
            if (move.MustCrossDangerZone(region) && !HasCrossedDangerZone(move))
            {
                return true;
            }
            else
            {
                _board.SucceedMove(move);
                return false;
            }
        }

        return false;
    }

    public bool HasCrossedDangerZone(Order order)
    {
        foreach (var battle in Battles)
        {
            if (battle.DangerZone is not null && battle.Results[0].Order == order)
            {
                return true;
            }
        }

        return false;
    }

    private void ResolveWinterOrders()
    {
        var allResolved = false;
        while (!allResolved)
        {
            allResolved = true;

            foreach (var (_, region) in _board.Regions)
            {
                // ReSharper disable once SwitchStatementMissingSomeEnumCasesNoDefault
                switch (region.Order?.Type)
                {
                    case OrderType.Build:
                        region.Unit = new Unit
                        {
                            Faction = region.Order.Faction,
                            Type = region.Order.UnitType,
                        };
                        region.Order = null;
                        break;
                    case OrderType.Disband:
                        region.RemoveUnit();
                        region.Order = null;
                        break;
                }

                if (!region.PartOfCycle)
                {
                    var cycle = _board.FindCycle(region.Name, region);
                    if (cycle is not null)
                    {
                        Board.PrepareCycleForResolving(cycle);
                    }
                }

                if (region.Order is not null)
                {
                    allResolved = false;
                    continue;
                }

                if (region.IncomingMoves.Count == 0)
                {
                    var move = region.IncomingMoves[0]; // Max 1 incoming move in winter
                    region.Unit = move.Unit();
                    _board.Regions[move.Origin].RemoveUnit();
                    _board.RemoveOrder(move);
                }
            }
        }
    }
}
