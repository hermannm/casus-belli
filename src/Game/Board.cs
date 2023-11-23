using System.Collections.Generic;
using System.Linq;

namespace Immerse.BfhClient.Game;

/// <summary>
/// Maps region names to regions.
/// </summary>
public class Board : Dictionary<string, Region>
{
    public void RemoveOrder(Order order)
    {
        var origin = this[order.Origin];
        origin.Order = null;

        if (order.Type == OrderType.Move)
        {
            this[order.Destination!].IncomingMoves.Remove(order);
        }
    }

    public bool AllResolved()
    {
        return Values.All(region => region.Resolved);
    }

    public (Region region2, bool sameFaction)? DiscoverTwoWayCycle(Region region1)
    {
        if (region1.Order?.Type != OrderType.Move)
        {
            return null;
        }

        var region2 = this[region1.Order.Destination!];
        if (region2.Order?.Type != OrderType.Move)
        {
            return null;
        }

        if (region1.Name != region2.Order.Destination!)
        {
            return null;
        }

        return (region2, region1.Order.Faction == region2.Order.Faction);
    }
}
