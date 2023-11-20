using System.Collections.Generic;

namespace Immerse.BfhClient.Game;

/// <summary>
/// Results of a battle between players, an attempt to conquer a neutral region, or an attempt to
/// cross a danger zone.
/// </summary>
public record Battle
{
    /// <summary>
    /// If length is one, the battle was a neutral region conquest attempt or danger zone crossing.
    /// If length is more than one, the battle was between players.
    /// </summary>
    public required List<Result> Results { get; set; }

    /// <summary>
    /// If battle was from a danger zone crossing: name of the danger zone.
    /// </summary>
    public string? DangerZone { get; set; }
}

/// <summary>
/// Dice and modifier result for a battle.
/// </summary>
public record Result
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
