using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// A typed number that adds to a player's result in a battle.
/// </summary>
public record struct Modifier
{
    /// <summary>
    /// The source of the modifier.
    /// Can only be of the constants defined in <see cref="ModifierType"/>.
    /// </summary>
    [JsonPropertyName("type")]
    public required string Type { get; set; }

    /// <summary>
    /// The positive or negative number that modifies the result total.
    /// </summary>
    [JsonPropertyName("value")]
    public required int Value { get; set; }

    /// <summary>
    /// If modifier was from a support: the supporting player.
    /// </summary>
    [JsonPropertyName("supportingPlayer")]
    public string? SupportingPlayer { get; set; }
}

/// <summary>
/// Valid values for a result modifier's type.
/// </summary>
public static class ModifierType
{
    /// <summary>
    /// Bonus from a random dice roll.
    /// </summary>
    public const string Dice = "dice";

    /// <summary>
    /// Bonus for the type of unit.
    /// </summary>
    public const string Unit = "unit";

    /// <summary>
    /// Penalty for attacking a neutral or defended forested area.
    /// </summary>
    public const string Forest = "forest";

    /// <summary>
    /// Penalty for attacking a neutral or defended castle area.
    /// </summary>
    public const string Castle = "castle";

    /// <summary>
    /// Penalty for attacking across a river, from the sea, or across a transport.
    /// </summary>
    public const string Water = "water";

    /// <summary>
    /// Bonus for attacking across a danger zone and surviving.
    /// </summary>
    public const string Surprise = "surprise";

    /// <summary>
    /// Bonus from supporting player in a battle.
    /// </summary>
    public const string Support = "support";
}
