using System.Collections.Generic;

namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// Dice and modifier result for a battle.
/// </summary>
public record struct Result
{
    public required int Total { get; set; }
    public required List<Modifier> Parts { get; set; }

    /// <summary>
    /// If result of a move order to the battle: the move order in question.
    /// </summary>
    public Order? Move { get; set; }

    /// <summary>
    /// If result of a defending unit in a region: the name of the region.
    /// </summary>
    public string? DefenderRegion { get; set; }
}
