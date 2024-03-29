namespace CasusBelli.Client.Game;

/// <summary>
/// A unit on the board, controlled by a player faction.
/// </summary>
public record Unit
{
    public required UnitType Type { get; set; }
    public required string Faction { get; set; }
}

/// <summary>
/// Valid values for a player unit's type.
/// </summary>
public enum UnitType
{
    /// <summary>
    /// A land unit that gets a +1 modifier in battle.
    /// </summary>
    Footman = 1,

    /// <summary>
    /// A land unit that moves 2 regions at a time.
    /// </summary>
    Knight,

    /// <summary>
    /// A unit that can move into sea regions and coastal regions.
    /// </summary>
    Ship,

    /// <summary>
    /// A land unit that instantly conquers neutral castles, and gets a +1 modifier in attacks on
    /// castles.
    /// </summary>
    Catapult
}
