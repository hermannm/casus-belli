using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// Dice and modifier result for a battle.
/// </summary>
public record struct Result
{
    /// <summary>
    /// The sum of the dice roll and modifiers.
    /// </summary>
    [JsonPropertyName("total")]
    public required int Total { get; set; }

    /// <summary>
    /// The modifiers comprising the result, including the dice roll.
    /// </summary>
    [JsonPropertyName("parts")]
    public required List<Modifier> Parts { get; set; }

    /// <summary>
    /// If result of a move order to the battle: the move order in question.
    /// </summary>
    [JsonPropertyName("move")]
    public Order? Move { get; set; }

    /// <summary>
    /// If result of a defending unit in a region: the name of the region.
    /// </summary>
    [JsonPropertyName("defenderRegion")]
    public string? DefenderRegion { get; set; }
}
