using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.Messages;

public record struct LobbyInfo
{
    public required string Name { get; set; }
    public required BoardInfo BoardInfo { get; set; }
}

public record struct BoardInfo
{
    [JsonPropertyName("ID")]
    public required string Id { get; set; }
    public required string Name { get; set; }
    public required int WinningCastleCount { get; set; }
}
