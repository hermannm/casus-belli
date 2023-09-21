using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.Messages;

public struct LobbyInfo
{
    [JsonPropertyName("lobbyName")]
    public required string LobbyName;

    [JsonPropertyName("gameName")]
    public required string GameName;
}
