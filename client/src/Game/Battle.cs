using System.Collections.Generic;
using System.Linq;
using Godot;

namespace CasusBelli.Client.Game;

/// <summary>
/// Results of a battle between players, an attempt to conquer a neutral region, or an attempt to
/// cross a danger zone.
/// </summary>
public partial class Battle : GodotObject
{
    /// <summary>
    /// If length is one, the battle was a neutral region conquest attempt or danger zone crossing.
    /// If length is more than one, the battle was between players.
    /// </summary>
    public required List<Result> Results { get; set; }

    /// <summary>
    /// If the battle is a danger zone crossing: name of the crossed danger zone.
    /// </summary>
    public string? DangerZone { get; set; }

    public List<string> RegionNames()
    {
        var regionNames = new List<string>();

        foreach (var result in Results)
        {
            if (result.Order is not null && !regionNames.Contains(result.Order.Destination!))
            {
                regionNames.Add(result.Order.Destination!);
            }
        }

        return regionNames;
    }

    public (bool isDangerZoneCrossing, bool succeeded, Order? order) IsDangerZoneCrossing()
    {
        if (DangerZone is null)
        {
            return (false, false, null);
        }

        const int minResultToSurviveDangerZone = 3;

        var result = Results[0];

        return (true, result.Total >= minResultToSurviveDangerZone, result.Order);
    }

    public bool IsBorderBattle()
    {
        if (Results.Count != 2)
        {
            return false;
        }

        var result1 = Results[0];
        var result2 = Results[1];
        if (result1.Order is null || result2.Order is null)
        {
            return false;
        }

        return result1.Order.Destination == result2.Order.Origin
            && result2.Order.Destination == result1.Order.Origin;
    }

    private const int MinResultToConquerNeutralRegion = 4;

    public (List<string> winnerFactions, List<string> loserFactions) WinnersAndLosers()
    {
        var winners = new List<string>();
        var losers = new List<string>();

        if (Results.Count == 1)
        {
            var result = Results[0];
            if (result.Total >= MinResultToConquerNeutralRegion)
            {
                winners.Add(result.Order!.Faction);
            }
            else
            {
                losers.Add(result.Order!.Faction);
            }
        }
        else
        {
            var highestResult = Results.Select(result => result.Total).Max();

            foreach (var result in Results)
            {
                var faction = result.DefenderFaction ?? result.Order!.Faction;
                if (result.Total >= highestResult)
                {
                    winners.Add(faction);
                }
                else
                {
                    losers.Add(faction);
                }
            }
        }

        return (winners, losers);
    }
}
