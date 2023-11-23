using System.Collections.Generic;

namespace Immerse.BfhClient.Game;

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
    /// If result of a defending unit in a region: the faction of the defender.
    /// </summary>
    public string? DefenderFaction { get; set; }
}
