using System.Collections.Generic;

namespace Immerse.BfhClient.Game;

/// <summary>
/// Maps region names to regions.
/// </summary>
public class Board : Dictionary<string, Region> { }

/// <summary>
/// A region on the board.
/// </summary>
public record Region
{
    public string Name { get; set; }
    public List<Neighbor> Neighbors { get; set; }

    /// <summary>
    /// Whether the region is a sea region that can only have ship units.
    /// </summary>
    public bool IsSea { get; set; }

    /// <summary>
    /// For land regions: affects the difficulty of conquering the region.
    /// </summary>
    public bool IsForest { get; set; }

    /// <summary>
    /// For land regions: affects the difficulty of conquering the region, and the points gained
    /// from it.
    /// </summary>
    public bool HasCastle { get; set; }

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
}

public record Neighbor
{
    public string Name { get; set; }

    /// <summary>
    /// Whether a river separates the neighboring regions, or this region is a sea and the neighbor
    /// is a land region.
    /// </summary>
    public bool IsAcrossWater { get; set; }

    /// <summary>
    /// Whether coast between neighboring land regions have cliffs (impassable to ships).
    /// </summary>
    public bool HasCliffs { get; set; }

    /// <summary>
    /// If not null: the name of the danger zone that the neighboring region lies across (requires
    /// check to pass).
    /// </summary>
    public string? DangerZone { get; set; }
}
