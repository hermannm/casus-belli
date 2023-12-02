using System.Collections.Generic;
using System.Linq;

namespace CasusBelli.Client.Game;

public record TransportPath(bool Attacked, string? DangerZone);

public static class TransportResolver
{
    /// <returns>Whether the region must wait to resolve.</returns>
    public static bool ResolveUncontestedTransports(Region region, Board board)
    {
        if (region.TransportsResolved)
        {
            return false;
        }

        foreach (var move in region.IncomingMoves)
        {
            var path = ResolveTransport(move, region, board);
            if (path is not null && (path.Attacked || path.DangerZone is not null))
            {
                return true;
            }
        }

        region.TransportsResolved = true;
        return false;
    }

    public static void ResolveContestedTransports(Region region, Board board)
    {
        if (region.TransportsResolved)
        {
            return;
        }

        foreach (var move in region.IncomingMoves)
        {
            ResolveTransport(move, region, board);
        }

        region.TransportsResolved = true;
    }

    private static TransportPath? ResolveTransport(Order move, Region destination, Board board)
    {
        if (destination.AdjacentTo(move.Origin))
        {
            return null;
        }

        var path = FindTransportPath(move.Origin, move.Destination!, board);
        if (path is null)
        {
            board.RetreatMove(move);
            return null;
        }

        return path;
    }

    public static TransportPath? FindTransportPath(
        string originName,
        string destinationName,
        Board board
    )
    {
        var origin = board.Regions[originName];
        if (origin.Empty() || origin.Unit?.Type == UnitType.Ship || origin.Sea)
        {
            return null;
        }

        return RecursivelyFindTransportPath(origin, destinationName, new HashSet<string>(), board);
    }

    private static TransportPath? RecursivelyFindTransportPath(
        Region region,
        string destination,
        HashSet<string> excludedRegions,
        Board board
    )
    {
        var (transportingNeighbors, newExcludedRegions) = GetTransportingNeighbors(
            region,
            board,
            excludedRegions
        );

        var paths = new List<TransportPath>();

        foreach (var transportNeighbor in transportingNeighbors)
        {
            var transportRegion = board.Regions[transportNeighbor.Name];

            var (destinationAdjacent, destinationDangerZone) = CheckNeighborsForDestination(
                transportRegion,
                destination
            );

            var nextPath = RecursivelyFindTransportPath(
                transportRegion,
                destination,
                newExcludedRegions,
                board
            );

            var subPaths = new List<TransportPath>();
            if (destinationAdjacent)
            {
                subPaths.Add(new TransportPath(transportRegion.Attacked(), destinationDangerZone));
            }
            if (nextPath is not null)
            {
                subPaths.Add(
                    nextPath with
                    {
                        Attacked = transportRegion.Attacked() || nextPath.Attacked
                    }
                );
            }

            var bestSubPath = BestTransportPath(subPaths);
            if (bestSubPath is not null)
            {
                paths.Add(bestSubPath);
            }
        }

        return BestTransportPath(paths);
    }

    private static (
        List<Neighbor> transports,
        HashSet<string> newExcludedRegions
    ) GetTransportingNeighbors(Region region, Board board, HashSet<string> excludedRegions)
    {
        var transports = new List<Neighbor>();
        var newExcludedRegions = new HashSet<string>(excludedRegions);

        if (region.Empty())
        {
            return (transports, newExcludedRegions);
        }

        foreach (var neighbor in region.Neighbors)
        {
            var neighborRegion = board.Regions[neighbor.Name];

            if (
                excludedRegions.Contains(neighbor.Name)
                || neighborRegion.Order?.Type != OrderType.Transport
                || neighborRegion.Unit?.Faction != region.Unit?.Faction
            )
            {
                continue;
            }

            transports.Add(neighbor);
            newExcludedRegions.Add(neighbor.Name);
        }

        return (transports, newExcludedRegions);
    }

    private static (
        bool destinationAdjacent,
        string? destinationDangerZone
    ) CheckNeighborsForDestination(Region region, string destination)
    {
        var destinationAdjacent = false;
        string? destinationDangerZone = null;

        foreach (var neighbor in region.Neighbors)
        {
            if (neighbor.Name == destination)
            {
                if (destinationAdjacent)
                {
                    if (destinationDangerZone is not null)
                    {
                        destinationDangerZone = neighbor.DangerZone;
                    }
                    continue;
                }

                destinationAdjacent = true;
                destinationDangerZone = neighbor.DangerZone;
            }
        }

        return (destinationAdjacent, destinationDangerZone);
    }

    private static TransportPath? BestTransportPath(List<TransportPath> paths)
    {
        if (paths.Count == 0)
        {
            return null;
        }

        var bestPath = paths[0];

        foreach (var path in paths.Skip(1))
        {
            if (bestPath.Attacked)
            {
                if (path.Attacked)
                {
                    if (path.DangerZone is null && bestPath.DangerZone is not null)
                    {
                        bestPath = path;
                    }
                }
                else
                {
                    bestPath = path;
                }
            }
            else if (!path.Attacked)
            {
                if (path.DangerZone is null && bestPath.DangerZone is not null)
                {
                    bestPath = path;
                }
            }
        }

        return bestPath;
    }
}
