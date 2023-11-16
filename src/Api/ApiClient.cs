using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;
using Immerse.BfhClient.Api.Messages;
using Godot;
using System.Net.Http.Json;
using Immerse.BfhClient.UI;
using HttpClient = System.Net.Http.HttpClient;
using GodotArray = Godot.Collections.Array;
using GodotDictionary = Godot.Collections.Dictionary;

namespace Immerse.BfhClient.Api;

/// <summary>
/// WebSocket client that connects to the game server.
/// Provides methods for sending and receiving messages to and from the server.
/// </summary>
public partial class ApiClient : Node
{
    /// ApiClient singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static ApiClient Instance { get; private set; } = null!;

    public Uri? ServerUrl { get; private set; }
    public LobbyInfo? Lobby { get; private set; }
    public bool HasJoinedLobby => Lobby != null;

    private readonly HttpClient _httpClient = new();
    private readonly ClientWebSocket _websocket = new();
    private readonly CancellationTokenSource _cancellation = new();
    private readonly MessageSender _messageSender;
    private readonly MessageReceiver _messageReceiver;
    private readonly Dictionary<MessageTag, StringName> _messageSignalNames;

    public ApiClient()
    {
        _websocket.Options.CollectHttpResponseDetails = true;
        _messageSender = new MessageSender(_websocket);
        _messageReceiver = new MessageReceiver(_websocket);
        _messageSignalNames = MessageDictionary.ReceivableMessageTags.Keys.ToDictionary(
            tag => tag,
            tag => new StringName(tag + "MessageReceived")
        );
    }

    public override void _EnterTree()
    {
        Instance = this;
        AddMessageReceivedSignals();
        AddMessageHandler<ErrorMessage>(DisplayServerError);

        if (OS.IsDebugBuild())
            TryConnect("localhost:8000");
    }

    public override void _Process(double delta)
    {
        if (!HasJoinedLobby)
            return;

        if (!_messageReceiver.MessageQueue.TryDequeue(out var message))
            return;

        if (!_messageSignalNames.TryGetValue(message.Tag, out var signal))
        {
            GD.PushError($"Received message tag '{message.Tag}' was not registered as signal");
            return;
        }

        var err = EmitSignal(signal, message.Data);
        if (err != Error.Ok)
            GD.PushError($"Failed to emit signal '{signal}': {err}");
    }

    public bool TryConnect(string serverUrl)
    {
        Uri parsedUrl;
        try
        {
            if (!serverUrl.StartsWith("http"))
            {
                serverUrl = $"http://{serverUrl}";
            }
            parsedUrl = new Uri(serverUrl, UriKind.Absolute);
        }
        catch (Exception e)
        {
            MessageDisplay.Instance.ShowError("Failed to parse given server URL", e.Message);
            return false;
        }

        ServerUrl = parsedUrl;
        _httpClient.BaseAddress = ServerUrl;
        return true;
    }

    /// <summary>
    /// Disconnects the API client from the server, and stops sending and receiving messages.
    /// </summary>
    public Task Disconnect()
    {
        ServerUrl = null;
        _httpClient.BaseAddress = null;
        return HasJoinedLobby ? LeaveLobby() : Task.CompletedTask;
    }

    public void SendMessage<TMessage>(TMessage message)
        where TMessage : GodotObject, ISendableMessage
    {
        _messageSender.SendQueue.Add(message);
    }

    public void AddMessageHandler<TMessage>(Action<TMessage> handler)
        where TMessage : GodotObject, IReceivableMessage
    {
        if (
            !MessageDictionary.ReceivableMessageTypes.TryGetValue(
                typeof(TMessage),
                out var messageTag
            )
        )
        {
            GD.PushError($"Invalid message type {typeof(TMessage)} for server message handler");
            return;
        }

        var signal = _messageSignalNames[messageTag];
        var err = Connect(signal, Callable.From(handler));
        if (err != Error.Ok)
            GD.PushError($"Failed to connect to signal '{signal}': {err}");
    }

    public async Task<List<LobbyInfo>?> ListLobbies()
    {
        List<LobbyInfo>? lobbies;
        try
        {
            lobbies = await _httpClient.GetFromJsonAsync<List<LobbyInfo>>("/lobbies");
        }
        catch (Exception e)
        {
            MessageDisplay.Instance.ShowError("Failed to get lobby list", e.Message);
            return null;
        }

        if (lobbies is null)
        {
            MessageDisplay.Instance.ShowError("Failed to get lobby list", "Response was empty");
            return null;
        }

        return lobbies;
    }

    /// <summary>
    /// Connects the API client to a server at the given URI, and starts sending and receiving
    /// messages.
    /// </summary>
    public async Task<bool> TryJoinLobby(LobbyInfo lobby, string username)
    {
        if (ServerUrl is null)
        {
            MessageDisplay.Instance.ShowError("Tried to join lobby before setting server URL");
            return false;
        }

        new Thread(() => _messageReceiver.ReadMessagesIntoQueue(_cancellation.Token)).Start();
        new Thread(() => _messageSender.SendMessagesFromQueue(_cancellation.Token)).Start();

        var joinLobbyUrl = new UriBuilder
        {
            Scheme = ServerUrl.Scheme == "https" ? "wss" : "ws",
            Host = ServerUrl.Host,
            Port = ServerUrl.Port,
            Path = "/join",
            Query = $"lobbyName={lobby.Name}&username={username}"
        };

        try
        {
            await _websocket.ConnectAsync(joinLobbyUrl.Uri, _httpClient, _cancellation.Token);
            _websocket.HttpResponseHeaders = null; // Frees now-redundant memory
        }
        catch (Exception e)
        {
            _cancellation.Cancel();

            string? errorMessage = null;
            // Since .NET ClientWebSockets do not provide us the HTTP response message in case of
            // failure, the server instead sends the error message through response headers
            if (_websocket.HttpResponseHeaders?.TryGetValue("Error", out var values) == true)
            {
                errorMessage = values.FirstOrDefault();
            }
            MessageDisplay.Instance.ShowError(
                "Failed to create WebSocket connection to server",
                errorMessage ?? e.Message
            );

            return false;
        }

        Lobby = lobby;
        return true;
    }

    public Task LeaveLobby()
    {
        _cancellation.Cancel();
        Lobby = null;

        return _websocket.CloseAsync(
            WebSocketCloseStatus.NormalClosure,
            "Client initiated disconnect from game server",
            _cancellation.Token
        );
    }

    private void AddMessageReceivedSignals()
    {
        foreach (var signal in _messageSignalNames.Values)
        {
            AddUserSignal(
                signal,
                new GodotArray
                {
                    new GodotDictionary
                    {
                        { "name", "message" },
                        { "type", (int)Variant.Type.Object }
                    }
                }
            );
        }
    }

    private static void DisplayServerError(ErrorMessage errorMessage)
    {
        MessageDisplay.Instance.ShowError(errorMessage.Error);
    }
}
