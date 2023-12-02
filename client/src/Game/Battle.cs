using System.Collections.Generic;

namespace CasusBelli.Client.Game;

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
    /// If the battle is a danger zone crossing: name of the crossed danger zone.
    /// </summary>
    public string? DangerZone { get; set; }
}
