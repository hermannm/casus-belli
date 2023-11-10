namespace Immerse.BfhClient.Api.Messages;

public record struct LobbyInfo
{
    public required string LobbyName { get; set; }
    public required string GameName { get; set; }
}
