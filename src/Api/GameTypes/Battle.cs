using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// Results of a battle from conflicting move orders, an attempt to conquer a neutral area,
/// or an attempt to cross a danger zone.
/// </summary>
public record struct Battle
{
    /// <summary>
    /// The dice and modifier results of the battle.
    /// If length is one, the battle was a neutral conquer attempt.
    /// If length is more than one, the battle was between players.
    /// </summary>
    [JsonPropertyName("results")]
    public required List<Result> Results { get; set; }

    /// <summary>
    /// In case of danger zone crossing: name of the danger zone.
    /// </summary>
    [JsonPropertyName("dangerZone")]
    public string? DangerZone { get; set; }
}
