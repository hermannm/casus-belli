using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// A player unit on the board.
/// </summary>
public record struct Unit
{
    /// <summary>
    /// Affects how the unit moves and its battle capabilities.
    /// Can only be of the constants defined in <see cref="UnitType"/>.
    /// </summary>
    [JsonPropertyName("type")]
    public required string Type { get; set; }

    /// <summary>
    /// The player owning the unit.
    /// </summary>
    [JsonPropertyName("player")]
    public required string Player { get; set; }
}

/// <summary>
/// Valid values for a player unit's type.
/// </summary>
public static class UnitType
{
    /// <summary>
    /// A land unit that gets a +1 modifier in battle.
    /// </summary>
    public const string Footman = "footman";

    /// <summary>
    /// A land unit that moves 2 regions at a time.
    /// </summary>
    public const string Horse = "horse";

    /// <summary>
    /// A unit that can move into sea regions and coastal regions.
    /// </summary>
    public const string Ship = "ship";

    /// <summary>
    /// A land unit that instantly conquers neutral castles, and gets a +1 modifier in attacks on
    /// castles.
    /// </summary>
    public const string Catapult = "catapult";
}
