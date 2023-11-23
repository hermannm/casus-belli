using System;
using System.Collections.Generic;
using Godot;
using Immerse.BfhClient.Api;
using Immerse.BfhClient.Api.Messages;
using Immerse.BfhClient.Lobby;
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

    private readonly Board _board = new();

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
        _board.Regions = message.Board;
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
        OrdersByFaction = message.OrdersByFaction;
        _board.PlaceOrders(OrdersByFaction, SupportCutSignal);

        Phase = GamePhase.ResolvingOrders;
        PhaseChangedSignal.Emit();

        ResolveMoves();
    }

    private void HandleBattleResultsMessage(BattleResultsMessage message)
    {
        throw new NotImplementedException();
    }

    private void ResolveMoves()
    {
        var allRegionsWaiting = false;
        while (!allRegionsWaiting)
        {
            allRegionsWaiting = true;

            foreach (var (_, region) in _board!.Regions)
            {
                if (region.Empty() && region.Controlled() && region.IncomingMoves.Count == 1)
                {
                    var move = region.IncomingMoves[0];
                    _board.Regions[move.Origin].MoveUnitTo(region);

                    UncontestedMoveSignal.Emit(
                        new UncontestedMove { FromRegion = move.Origin, ToRegion = region.Name }
                    );

                    allRegionsWaiting = true;
                }
            }
        }
    }
}
