using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.Messages;

public record struct LobbyInfo
{
    [JsonPropertyName("lobbyName")]
    public required string LobbyName { get; set; }

    [JsonPropertyName("gameName")]
    public required string GameName { get; set; }
}
