using System.Collections.Generic;
using System.Linq;
using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Game;

/// <summary>
/// Maps region names to regions.
/// </summary>
public class Board : Dictionary<string, Region>
{
    public void RemoveOrder(Order order)
    {
        var origin = this[order.Origin];
        origin.Order = null;

        if (order.Type == OrderType.Move)
        {
            this[order.Destination!].IncomingMoves.Remove(order);
        }
    }

    public bool AllResolved()
    {
        return Values.All(region => region.Resolved);
    }
}

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
    public required bool IsSea { get; set; }

    /// <summary>
    /// For land regions: affects the difficulty of conquering the region.
    /// </summary>
    public required bool IsForest { get; set; }

    /// <summary>
    /// For land regions: affects the difficulty of conquering the region, and the points gained
    /// from it.
    /// </summary>
    public required bool HasCastle { get; set; }

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
    public int? SiegeCount { get; set; }

    [JsonIgnore]
    public Order? Order { get; set; }

    [JsonIgnore]
    public List<Order> IncomingMoves { get; set; } = new();

    [JsonIgnore]
    public bool Resolved { get; set; } = false;

    public bool IsAttacked()
    {
        return IncomingMoves.Count > 0;
    }

    public bool IsEmpty()
    {
        return Unit == null;
    }

    public bool IsControlled()
    {
        return ControllingFaction != null;
    }

    public bool IsAdjacentTo(string regionName)
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

    public void MoveUnitTo(Region destination)
    {
        destination.Unit = Unit;
        Unit = null;
    }

    public void RemoveUnit(Unit unit)
    {
        if (unit == Unit)
        {
            Unit = null;
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
    public required bool IsAcrossWater { get; set; }

    /// <summary>
    /// Whether coast between neighboring land regions have cliffs (impassable to ships).
    /// </summary>
    public required bool HasCliffs { get; set; }

    /// <summary>
    /// If not null: the name of the danger zone that the neighboring region lies across (requires
    /// check to pass).
    /// </summary>
    public string? DangerZone { get; set; }
}
