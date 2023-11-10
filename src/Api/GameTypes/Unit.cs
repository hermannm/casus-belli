namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// A unit on the board, controlled by a player faction.
/// </summary>
public record struct Unit
{
    public required string Type { get; set; }
    public required string Faction { get; set; }
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
