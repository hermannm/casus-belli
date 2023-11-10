namespace Immerse.BfhClient.Api.GameTypes;

/// <summary>
/// An order submitted by a player for one of their units in a given round.
/// </summary>
public record struct Order
{
    public required string Type { get; set; }
    public required string Faction { get; set; }

    /// <summary>
    /// Name of the region where the order is placed.
    /// </summary>
    public required string Origin { get; set; }

    /// <summary>
    /// For move and support orders: name of destination region.
    /// </summary>
    public string? Destination { get; set; }

    /// <summary>
    /// For move orders with horse units: optional name of second destination region to move to if
    /// the first destination was reached.
    /// </summary>
    public string? SecondDestination { get; set; }

    /// <summary>
    /// For move orders: name of DangerZone the order tries to pass through, if any.
    /// </summary>
    public string? ViaDangerZone { get; set; }

    /// <summary>
    /// For build orders: type of unit to build.
    /// Can only be of the constants defined in <see cref="UnitType"/>.
    /// </summary>
    public string? Build { get; set; }
}

/// <summary>
/// Valid values for a player-submitted order's type.
/// </summary>
public static class OrderType
{
    /// <summary>
    /// An order for a unit to move from one area to another.
    /// Includes internal moves in winter.
    /// </summary>
    public const string Move = "move";

    /// <summary>
    /// An order for a unit to support battle in an adjacent area.
    /// </summary>
    public const string Support = "support";

    /// <summary>
    /// For ship unit at sea: an order to transport a land unit across the sea.
    /// </summary>
    public const string Transport = "transport";

    /// <summary>
    /// For land unit in unconquered castle area: an order to besiege the castle.
    /// </summary>
    public const string Besiege = "besiege";

    /// <summary>
    /// For player-controlled area in winter: an order for what type of unit to build in the area.
    /// </summary>
    public const string Build = "build";
}
