using System.Text.Json.Serialization;

namespace Immerse.BfhClient.Api.Messages;

/// <summary>
/// Messages sent between the game client and server look like this:
/// <code>
/// {
///     "tag": 4,
///     "data": {"gameId": "green"}
/// }
/// </code>
/// ...where the "tag" field is one of the enum values defined in <see cref="MessageTag"/>, and
/// "data" is one of the message structs in <see cref="Immerse.BfhClient.Api.Messages"/>.
/// </summary>
public record struct Message
{
    [JsonPropertyName("tag")]
    public required MessageTag Tag;

    [JsonPropertyName("data")]
    public required object Data;
}
