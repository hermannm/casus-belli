using System.Collections.Generic;
using System.Linq;
using System.Text.Json.Serialization;

namespace CasusBelli.Client.Game;

/// <summary>
/// A region on the board.
/// </summary>
public record Region
{
    public required string Name { get; set; }
    public required List<Neighbor> Neighbors { get; set; }

    /// <summary>
    /// Whether the region is a sea region that can only have ship units.
    /// </summary>
    public required bool Sea { get; set; }

    /// <summary>
    /// For land regions: affects the difficulty of conquering the region.
    /// </summary>
    public required bool Forest { get; set; }

    /// <summary>
    /// For land regions: affects the difficulty of conquering the region, and the points gained
    /// from it.
    /// </summary>
    public required bool Castle { get; set; }

    /// <summary>
    /// For land regions: the collection of regions that the region belongs to (affects units gained
    /// from conquering).
    /// </summary>
    public string? Nation { get; set; }

    /// <summary>
    /// For land regions that are a starting region for a player faction.
    /// </summary>
    public string? HomeFaction { get; set; }

    /// <summary>
    /// The unit that currently occupies the region.
    /// </summary>
    public Unit? Unit { get; set; }

    /// <summary>
    /// The player faction that currently controls the region.
    /// </summary>
    public string? ControllingFaction { get; set; }

    /// <summary>
    /// For land regions with castles: the number of times an occupying unit has besieged the
    /// castle.
    /// </summary>
    public int SiegeCount { get; set; } = 0;

    [JsonIgnore]
    public Order? Order { get; set; } = null;

    [JsonIgnore]
    public List<Order> IncomingMoves { get; set; } = new();

    [JsonIgnore]
    public List<Order> IncomingSupports { get; set; } = new();

    [JsonIgnore]
    public List<Order> IncomingKnightMoves { get; set; } = new();

    [JsonIgnore]
    public int ExpectedKnightMoves { get; set; } = 0;

    [JsonIgnore]
    public bool ResolvingKnightMoves { get; set; } = false;

    [JsonIgnore]
    public bool Resolved { get; set; } = false;

    [JsonIgnore]
    public bool TransportsResolved { get; set; } = false;

    [JsonIgnore]
    public bool DangerZonesResolved { get; set; } = false;

    [JsonIgnore]
    public bool PartOfCycle { get; set; } = false;

    [JsonIgnore]
    public Order? UnresolvedRetreat { get; set; } = null;

    public void ResetResolvingState()
    {
        Order = null;
        IncomingMoves = new List<Order>();
        IncomingSupports = new List<Order>();
        IncomingKnightMoves = new List<Order>();
        ExpectedKnightMoves = 0;
        ResolvingKnightMoves = false;
        Resolved = false;
        TransportsResolved = false;
        DangerZonesResolved = false;
        PartOfCycle = false;
        UnresolvedRetreat = null;
    }

    public bool Attacked()
    {
        return IncomingMoves.Count > 0;
    }

    public bool Empty()
    {
        return Unit == null;
    }

    public bool Controlled()
    {
        return ControllingFaction != null;
    }

    public void RemoveUnit()
    {
        if (!PartOfCycle)
        {
            Unit = null;
            SiegeCount = 0;
        }
    }

    public void ReplaceUnit(Unit unit)
    {
        Unit = unit;
        SiegeCount = 0;
    }

    public void ResolveRetreat()
    {
        if (UnresolvedRetreat is not null)
        {
            if (Empty())
            {
                Unit = UnresolvedRetreat.Unit();
            }
            UnresolvedRetreat = null;
        }
    }

    public bool AdjacentTo(string regionName)
    {
        return Neighbors.Any(neighbor => neighbor.Name == regionName);
    }

    /// <summary>
    /// Returns a region's neighbor of the given name, if found.
    /// If the region has several neighbor relations to the region, returns the one matching the
    /// provided 'viaDangerZone' string.
    /// </summary>
    public Neighbor? GetNeighbor(string neighborName, string? viaDangerZone)
    {
        Neighbor? neighbor = null;

        foreach (var candidate in Neighbors)
        {
            if (neighborName != candidate.Name)
            {
                continue;
            }

            if (neighbor is null)
            {
                neighbor = candidate;
            }
            else if (candidate.DangerZone is not null && viaDangerZone == candidate.DangerZone)
            {
                neighbor = candidate;
            }
        }

        return neighbor;
    }
}

public record Neighbor
{
    public required string Name { get; set; }

    /// <summary>
    /// Whether a river separates the neighboring regions, or this region is a sea and the neighbor
    /// is a land region.
    /// </summary>
    public required bool AcrossWater { get; set; }

    /// <summary>
    /// Whether coast between neighboring land regions have cliffs (impassable to ships).
    /// </summary>
    public required bool Cliffs { get; set; }

    /// <summary>
    /// If not null: the name of the danger zone that the neighboring region lies across (requires
    /// check to pass).
    /// </summary>
    public string? DangerZone { get; set; }
}
