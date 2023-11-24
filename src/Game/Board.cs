using System.Collections.Generic;
using System.Linq;

namespace Immerse.BfhClient.Game;

public class Board
{
    /// <summary>
    /// Maps region names to regions.
    /// </summary>
    public Dictionary<string, Region> Regions { get; set; } = new();

    public void PlaceOrder(Order order)
    {
        var origin = Regions[order.Origin];
        origin.Order = order;

        if (order.Type != OrderType.Build)
        {
            order.UnitType = origin.Unit!.Type;
        }

        if (order.Type == OrderType.Move)
        {
            var destination = Regions[order.Destination!];
            destination.IncomingMoves.Add(order);
            if (order.HasSecondHorseMove())
            {
                destination.ExpectedSecondHorseMoves++;
            }
        }
    }

    public void RemoveOrder(Order order)
    {
        if (!order.Retreat)
        {
            Regions[order.Origin].Order = null;
        }

        if (order.Type == OrderType.Move)
        {
            Regions[order.Destination!].IncomingMoves.Remove(order);
        }
    }

    public void SucceedMove(Order move)
    {
        var destination = Regions[move.Destination!];

        destination.ReplaceUnit(move.Unit());
        destination.Order = null;

        // Seas cannot be controlled, and unconquered castles must be besieged first, unless the
        // attacking unit is a catapult
        if (
            !destination.Sea
            && (
                !destination.Castle
                || destination.Controlled()
                || move.UnitType == UnitType.Catapult
            )
        )
        {
            destination.ControllingFaction = move.Faction;
        }

        Regions[move.Origin].RemoveUnit();
        RemoveOrder(move);

        var secondHorseMove = move.TryGetSecondHorseMove();
        if (secondHorseMove is not null)
        {
            Regions[secondHorseMove.Destination!].IncomingSecondHorseMoves.Add(secondHorseMove);
        }
    }

    public void KillMove(Order move)
    {
        RemoveOrder(move);
        if (!move.Retreat)
        {
            Regions[move.Origin].RemoveUnit();

            if (move.HasSecondHorseMove())
            {
                Regions[move.SecondDestination!].ExpectedSecondHorseMoves--;
            }
        }
    }

    public void RetreatMove(Order move)
    {
        RemoveOrder(move);

        var origin = Regions[move.Origin];
        if (!origin.Attacked())
        {
            origin.Unit = move.Unit();
        }
        else if (origin.PartOfCycle)
        {
            origin.UnresolvedRetreat = move;
        }
        else
        {
            var retreat = move with
            {
                Retreat = true,
                Origin = move.Destination!,
                Destination = move.Origin,
                SecondDestination = null
            };

            origin.IncomingMoves.Add(retreat);
            origin.Order = null;
            origin.RemoveUnit();
        }

        if (move.HasSecondHorseMove())
        {
            Regions[move.SecondDestination!].ExpectedSecondHorseMoves--;
        }
    }

    public bool AllResolved()
    {
        return Regions.Values.All(region => region.Resolved);
    }

    public List<Region>? DiscoverCycle(string firstRegionName, Order? order)
    {
        if (order?.Type != OrderType.Move)
        {
            return null;
        }

        var destination = Regions[order.Destination!];
        if (destination.Name == firstRegionName)
        {
            return new List<Region> { destination };
        }

        var cycle = DiscoverCycle(firstRegionName, destination.Order);
        if (cycle is null)
        {
            return null;
        }

        cycle.Add(destination);
        return cycle;
    }

    public (Region? region2, bool sameFaction) DiscoverTwoWayCycle(Region region1)
    {
        if (region1.Order?.Type != OrderType.Move)
        {
            return (null, false);
        }

        var region2 = Regions[region1.Order.Destination!];
        if (region2.Order?.Type != OrderType.Move)
        {
            return (null, false);
        }

        if (region1.Name != region2.Order.Destination!)
        {
            return (null, false);
        }

        return (region2, region1.Order.Faction == region2.Order.Faction);
    }

    public void PrepareCycleForResolving(List<Region> cycle)
    {
        foreach (var region in cycle)
        {
            region.RemoveUnit();
            region.Order = null;
            region.PartOfCycle = true;
        }
    }
}
