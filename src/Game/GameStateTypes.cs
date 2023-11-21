using Godot;

namespace Immerse.BfhClient.Game;

public enum GamePhase
{
    SubmittingOrders,
    OrdersSubmitted,
    ResolvingOrders
}

public partial class SupportCut : GodotObject
{
    public required string RegionName;
}

public partial class UncontestedMove : GodotObject
{
    public required string FromRegion;
    public required string ToRegion;
}
