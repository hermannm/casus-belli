using System.Collections.Generic;

namespace CasusBelli.Client.Game;

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
    public Order? Order { get; set; }

    /// <summary>
    /// If result of a defending unit in a region: the faction of the defender.
    /// </summary>
    public string? DefenderFaction { get; set; }
}
