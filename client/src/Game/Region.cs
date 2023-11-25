using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Game;

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
    public int ExpectedSecondHorseMoves { get; set; } = 0;

    [JsonIgnore]
    public List<Order> IncomingSecondHorseMoves { get; set; } = new();

    [JsonIgnore]
    public bool Resolved { get; set; } = false;

    [JsonIgnore]
    public Order? UnresolvedRetreat { get; set; } = null;

    /// <summary>
    /// Whether the region is part of a cycle of move orders.
    /// </summary>
    [JsonIgnore]
    public bool PartOfCycle { get; set; } = false;

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

    public bool AdjacentTo(string regionName)
    {
        foreach (var neighbor in Neighbors)
        {
            if (neighbor.Name == regionName)
            {
                return true;
            }
        }
        return false;
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

    public void ResolveRetreatIfUnresolved()
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
