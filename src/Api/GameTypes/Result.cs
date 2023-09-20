using System.Collections.Generic;
using Newtonsoft.Json;

namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// Dice and modifier result for a battle.
/// </summary>
public struct Result
{
    /// <summary>
    /// The sum of the dice roll and modifiers.
    /// </summary>
    [JsonProperty("total", Required = Required.Always)]
    public int Total;

    /// <summary>
    /// The modifiers comprising the result, including the dice roll.
    /// </summary>
    [JsonProperty("parts", Required = Required.Always)]
    public List<Modifier> Parts;

    /// <summary>
    /// If result of a move order to the battle: the move order in question.
    /// </summary>
    [JsonProperty("move")]
    public Order? Move;

    /// <summary>
    /// If result of a defending unit in a region: the name of the region.
    /// </summary>
    [JsonProperty("defenderRegion")]
    public string? DefenderRegion;
}
