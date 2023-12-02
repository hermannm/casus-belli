namespace CasusBelli.Client.Game;

/// <summary>
/// An order submitted by a player for one of their units in a given round.
/// </summary>
public record Order
{
    public required OrderType Type { get; set; }

    /// <summary>
    /// For build orders: the type of unit moved.
    /// For all other orders: the type of unit in the ordered region.
    /// </summary>
    public required UnitType UnitType { get; set; }

    /// <summary>
    /// For move orders that lost a singleplayer battle or tied a multiplayer battle, and have to
    /// fight their way back to their origin region. Must be false when submitting orders.
    /// </summary>
    public bool Retreat { get; set; } = false;

    /// <summary>
    /// The faction of the player that submitted the order.
    /// </summary>
    public required string Faction { get; set; }

    /// <summary>
    /// The region where the order was placed.
    /// </summary>
    public required string Origin { get; set; }

    /// <summary>
    /// For move and support orders: name of destination region.
    /// </summary>
    public string? Destination { get; set; }

    /// <summary>
    /// For move orders with knight units: optional name of second destination region to move to if
    /// the first destination was reached.
    /// </summary>
    public string? SecondDestination { get; set; }

    /// <summary>
    /// For move orders: name of DangerZone the order tries to pass through, if any.
    /// </summary>
    public string? ViaDangerZone { get; set; }

    public Unit Unit()
    {
        return new Unit { Faction = Faction, Type = UnitType };
    }

    public bool HasKnightMove()
    {
        return Type == OrderType.Move
            && UnitType == UnitType.Knight
            && SecondDestination is not null;
    }

    public Order? TryGetKnightMove()
    {
        if (!HasKnightMove())
        {
            return null;
        }

        return this with
        {
            Origin = Destination!,
            Destination = SecondDestination,
            SecondDestination = null
        };
    }

    public bool MustCrossDangerZone(Region destination)
    {
        var neighbor = destination.GetNeighbor(Origin, ViaDangerZone);
        return neighbor?.DangerZone != null;
    }
}

/// <summary>
/// Valid values for a player-submitted order's type.
/// </summary>
public enum OrderType
{
    /// <summary>
    /// An order for a unit to move from one area to another.
    /// Includes internal moves in winter.
    /// </summary>
    Move = 1,

    /// <summary>
    /// An order for a unit to support battle in an adjacent area.
    /// </summary>
    Support,

    /// <summary>
    /// For ship unit at sea: an order to transport a land unit across the sea.
    /// </summary>
    Transport,

    /// <summary>
    /// For land unit in unconquered castle area: an order to besiege the castle.
    /// </summary>
    Besiege,

    /// <summary>
    /// For player-controlled area in winter: an order for what type of unit to build in the area.
    /// </summary>
    Build,

    /// <summary>
    /// For region with a player's own unit, in winter: an order to disband the unit, when their
    /// current number of units exceeds their max number of units.
    /// </summary>
    Disband
}
