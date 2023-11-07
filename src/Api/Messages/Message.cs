using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Messages sent between the game client and server look like this:
/// <code>
/// {
///     "type": 4,
///     "data": {"gameId": "green"}
/// }
/// </code>
/// ...where the "type" field is one of the enum values defined in <see cref="MessageType"/>, and
/// "data" is one of the message structs in <see cref="Immerse.BfhClient.Api.Messages"/>.
/// </summary>
public record struct Message
{
    [JsonPropertyName("type")]
    public required MessageType Type;

    [JsonPropertyName("data")]
    public required object Data;
}
