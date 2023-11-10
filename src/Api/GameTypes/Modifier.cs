namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// A typed number that adds to a player's result in a battle.
/// </summary>
public record struct Modifier
{
    public required ModifierType Type { get; set; }
    public required int Value { get; set; }

    /// <summary>
    /// Non-null if Type is Support.
    /// </summary>
    public string? SupportingPlayer { get; set; }
}

public enum ModifierType
{
    /// <summary>
    /// Bonus from a random dice roll.
    /// </summary>
    Dice = 1,

    /// <summary>
    /// Bonus for the type of unit.
    /// </summary>
    Unit,

    /// <summary>
    /// Penalty for attacking a neutral or defended forested area.
    /// </summary>
    Forest,

    /// <summary>
    /// Penalty for attacking a neutral or defended castle area.
    /// </summary>
    Castle,

    /// <summary>
    /// Penalty for attacking across a river, from the sea, or across a transport.
    /// </summary>
    Water,

    /// <summary>
    /// Bonus for attacking across a danger zone and surviving.
    /// </summary>
    Surprise,

    /// <summary>
    /// Bonus from supporting player in a battle.
    /// </summary>
    Support
}
