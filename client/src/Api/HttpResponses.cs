using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace CasusBelli.Client.Api;

public record LobbyInfo
{
    public required string Name { get; set; }
    public required int PlayerCount { get; set; }
    public required BoardInfo BoardInfo { get; set; }
}

public record BoardInfo
{
    [JsonPropertyName("ID")]
    public required string Id { get; set; }
    public required string Name { get; set; }
    public required int WinningCastleCount { get; set; }
    public required List<string> PlayerFactions { get; set; }
}
