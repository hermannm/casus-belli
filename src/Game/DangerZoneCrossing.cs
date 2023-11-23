namespace Immerse.BfhClient.Game;

/// <summary>
/// Result of an order that had to cross a danger zone to its destination, rolling dice to succeed.
/// For move orders, the moved unit dies if it fails the crossing.
/// For support orders, the support is cut if it fails the crossing.
/// </summary>
public class DangerZoneCrossing
{
    public required string DangerZone { get; set; }
    public required bool Survived { get; set; }
    public required int DiceResult { get; set; }
    public required Order Order { get; set; }
}
