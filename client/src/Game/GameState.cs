using System.Collections.Generic;
using CasusBelli.Client.Api;
using CasusBelli.Client.Lobby;
using Godot;

namespace CasusBelli.Client.Game;

public enum GamePhase
{
    SubmittingOrders,
    OrdersSubmitted,
    ResolvingOrders
}

public partial class GameState : Node
{
    public static GameState Instance { get; private set; } = null!;

    public Season Season { get; private set; } = Season.Winter;
    public Dictionary<string, List<Order>> OrdersByFaction { get; private set; } = new();
    public List<Player> PlayersYetToSubmitOrders = new();
    public List<Battle> Battles = new();
    public Battle? CurrentBattle;

    private GamePhase _phase = GamePhase.SubmittingOrders;
    public GamePhase Phase
    {
        get => _phase;
        private set
        {
            _phase = value;
            EmitSignal(SignalName.PhaseChanged, (int)_phase);
        }
    }

    [Signal]
    public delegate void PhaseChangedEventHandler(GamePhase phase);

    [Signal]
    public delegate void SupportCutEventHandler(string regionName);

    [Signal]
    public delegate void UncontestedMoveEventHandler(string fromRegion, string toRegion);

    [Signal]
    public delegate void BattleAnnouncementEventHandler(Battle battle);

    private readonly Board _board = new();

    public override void _EnterTree()
    {
        Instance = this;
        ApiClient.Instance.AddMessageHandler<GameStartedMessage>(HandleGameStarted);
        ApiClient.Instance.AddMessageHandler<OrderRequestMessage>(HandleOrderRequest);
        ApiClient.Instance.AddMessageHandler<OrdersConfirmationMessage>(HandleOrdersConfirmation);
        ApiClient.Instance.AddMessageHandler<OrdersReceivedMessage>(HandleOrdersReceived);
        ApiClient.Instance.AddMessageHandler<BattleAnnouncementMessage>(HandleBattleAnnouncement);
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
        Phase = GamePhase.SubmittingOrders;
        OrdersByFaction.Clear();
        PlayersYetToSubmitOrders = new List<Player>(LobbyState.Instance.OtherPlayers);
    }

    private void HandleOrdersConfirmation(OrdersConfirmationMessage message)
    {
        if (message.FactionThatSubmittedOrders == LobbyState.Instance.Player.Faction)
        {
            Phase = GamePhase.OrdersSubmitted;
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
        Phase = GamePhase.ResolvingOrders;

        if (Season == Season.Winter)
        {
            ResolveWinterOrders();
        }
        else
        {
            ResolveUncontestedRegions();
        }
    }

    private void HandleBattleAnnouncement(BattleAnnouncementMessage message)
    {
        EmitSignal(SignalName.BattleAnnouncement, message.Battle);
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
        if (isDangerZoneCrossing)
        {
            if (!succeeded)
            {
                if (order!.Type == OrderType.Move)
                {
                    _board.KillMove(order);
                }
                else
                {
                    _board.RemoveOrder(order);
                }
            }
        }
        else if (battle.IsBorderBattle())
        {
            ResolveBorderBattle(battle);
        }
        else if (battle.Results.Count == 1)
        {
            ResolveSingleplayerBattle(battle);
        }
        else
        {
            ResolveMultiplayerBattle(battle);
        }
    }

    private void ResolveSingleplayerBattle(Battle battle)
    {
        var move = battle.Results[0].Order!;

        var (winners, _) = battle.WinnersAndLosers();
        if (winners.Count == 1)
        {
            _board.SucceedMove(move);
        }
        else
        {
            _board.RetreatMove(move);
        }
    }

    private void ResolveMultiplayerBattle(Battle battle)
    {
        var (winners, losers) = battle.WinnersAndLosers();
        var tie = winners.Count > 1;

        foreach (var result in battle.Results)
        {
            if (result.DefenderFaction is not null)
            {
                // If the defender won or or was part of a tie, nothing changes for them.
                // If an attacker won alone and the defender controlled the region, the defender
                // will be removed as part of succeedMove for the winner.
                // If the defender was on the losing end of a tie in a battle with multiple
                // combatants, or the defender lost but did not control the region, we have to
                // remove the unit here.
                if (losers.Contains(result.DefenderFaction))
                {
                    // Guaranteed to have 1 element, since this is not a border battle
                    var regionName = battle.RegionNames()[0];
                    var region = _board.Regions[regionName];
                    if (tie || !region.Controlled())
                    {
                        region.RemoveUnit();
                    }
                }

                continue;
            }

            var move = result.Order!;
            if (losers.Contains(move.Faction))
            {
                _board.KillMove(move);
                continue;
            }

            if (tie)
            {
                _board.RetreatMove(move);
                continue;
            }

            // If the destination is not controlled, then the winner will have to battle there
            // before we can succeed the move
            if (_board.Regions[move.Destination!].Controlled())
            {
                _board.SucceedMove(move);
            }
        }
    }

    private void ResolveBorderBattle(Battle battle)
    {
        var (winners, losers) = battle.WinnersAndLosers();

        // If battle was a tie, both moves retreat
        if (winners.Count > 1)
        {
            var order1 = battle.Results[0].Order!;
            var order2 = battle.Results[1].Order!;

            // Remove both orders before retreating, so they don't think their origins are attacked
            _board.RemoveOrder(order1);
            _board.RemoveOrder(order2);

            _board.RetreatMove(order1);
            _board.RetreatMove(order2);
        }
        else
        {
            foreach (var result in battle.Results)
            {
                // Only the loser is affected by the results of the border battle; the winner may
                // still have to win a battle in the destination region, which will be handled by
                // the next cycle of move resolving.
                if (result.Order!.Faction == losers[0])
                {
                    _board.KillMove(result.Order);
                    break;
                }
            }
        }
    }

    private void ResolveUncontestedRegions()
    {
        var allRegionsWaiting = false;
        while (!allRegionsWaiting)
        {
            allRegionsWaiting = true;

            foreach (var (_, region) in _board.Regions)
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
