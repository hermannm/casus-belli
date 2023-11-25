using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http.Json;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;
using CasusBelli.Client.Api.Messages;
using CasusBelli.Client.UI;
using Godot;
using HttpClient = System.Net.Http.HttpClient;
using GodotArray = Godot.Collections.Array;
using GodotDictionary = Godot.Collections.Dictionary;

namespace CasusBelli.Client.Api;

/// <summary>
/// WebSocket client that connects to the game server.
/// Provides methods for sending and receiving messages to and from the server.
/// </summary>
public partial class ApiClient : Node
{
    /// ApiClient singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static ApiClient Instance { get; private set; } = null!;

    public Uri? ServerUrl => _httpClient?.BaseAddress;

    private HttpClient? _httpClient = null;
    private ClientWebSocket? _socket = null;
    private CancellationTokenSource? _cancellation = null;
    private readonly MessageSender _messageSender = new();
    private readonly MessageReceiver _messageReceiver = new();
    private readonly Dictionary<MessageTag, StringName> _messageSignalNames =
        MessageDictionary.ReceivableMessageTags.Keys.ToDictionary(
            tag => tag,
            tag => new StringName(tag + "MessageReceived")
        );
    private bool _hasJoinedLobby = false;

    public override void _EnterTree()
    {
        Instance = this;
        AddMessageReceivedSignals();
        AddMessageHandler<ErrorMessage>(HandleServerError);

        if (OS.IsDebugBuild())
        {
            TryConnect("localhost:8000");
        }
    }

    public override void _Process(double delta)
    {
        if (!_hasJoinedLobby)
        {
            return;
        }
        if (!_messageReceiver.MessageQueue.TryDequeue(out var message))
        {
            return;
        }
        if (!_messageSignalNames.TryGetValue(message.Tag, out var signal))
        {
            GD.PushError($"Received message tag '{message.Tag}' was not registered as signal");
            return;
        }

        var err = EmitSignal(signal, message.Data);
        if (err != Error.Ok)
        {
            GD.PushError($"Failed to emit signal '{signal}': {err}");
        }
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

        _httpClient = new HttpClient();
        _httpClient.BaseAddress = parsedUrl;
        return true;
    }

    /// <summary>
    /// Disconnects the API client from the server, and stops sending and receiving messages.
    /// </summary>
    public Task Disconnect()
    {
        _httpClient = null;
        return _hasJoinedLobby ? LeaveLobby() : Task.CompletedTask;
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
        var error = Connect(signal, Callable.From(handler));
        if (error != Error.Ok)
        {
            GD.PushError($"Failed to connect signal '{signal}': {error}");
        }
    }

    public async Task<List<LobbyInfo>?> ListLobbies()
    {
        List<LobbyInfo>? lobbies;
        try
        {
            lobbies = await _httpClient!.GetFromJsonAsync<List<LobbyInfo>>("/lobbies");
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
    public async Task<bool> TryJoinLobby(string lobbyName, string username)
    {
        if (ServerUrl is null)
        {
            MessageDisplay.Instance.ShowError("Tried to join lobby before setting server URL");
            return false;
        }

        // We have to create a new ClientWebSocket here, as it does not support using the same
        // instance across connections
        _socket = new ClientWebSocket();
        // This will populate the socket's HttpResponseHeaders, which we use for error messages
        _socket.Options.CollectHttpResponseDetails = true;

        // Cannot be reused, so must be recreated for each lobby joined
        _cancellation = new CancellationTokenSource();

        new Thread(
            () => _messageReceiver.ReadMessagesIntoQueue(_socket, _cancellation.Token)
        ).Start();
        new Thread(
            () => _messageSender.SendMessagesFromQueue(_socket, _cancellation.Token)
        ).Start();

        var joinLobbyUrl = new UriBuilder
        {
            Scheme = ServerUrl.Scheme == "https" ? "wss" : "ws",
            Host = ServerUrl.Host,
            Port = ServerUrl.Port,
            Path = "/join",
            Query = $"lobbyName={lobbyName}&username={username}"
        };

        try
        {
            await _socket.ConnectAsync(joinLobbyUrl.Uri, _httpClient, _cancellation.Token);
            _socket.HttpResponseHeaders = null; // Frees now-redundant memory
        }
        catch (Exception e)
        {
            _cancellation.Cancel();

            string? errorMessage = null;
            // Since .NET ClientWebSockets do not provide us the HTTP response message in case of
            // failure, the server instead sends the error message through response headers
            if (_socket.HttpResponseHeaders?.TryGetValue("Error", out var values) == true)
            {
                errorMessage = values.FirstOrDefault();
            }
            MessageDisplay.Instance.ShowError(
                "Failed to create WebSocket connection to server",
                errorMessage ?? e.Message
            );

            return false;
        }

        _hasJoinedLobby = true;
        return true;
    }

    public async Task LeaveLobby()
    {
        _cancellation?.Cancel();
        _hasJoinedLobby = false;
        _messageReceiver.ClearQueue();
        _messageSender.ClearQueue();
        if (_socket is null)
        {
            return;
        }

        try
        {
            await _socket.CloseAsync(
                WebSocketCloseStatus.NormalClosure,
                "client initiated disconnect from game server",
                new CancellationToken()
            );
        }
        catch (Exception e)
        {
            GD.PushError($"Closing WebSocket connection: {e}");
        }
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

    private static void HandleServerError(ErrorMessage errorMessage)
    {
        MessageDisplay.Instance.ShowError(errorMessage.Error);
    }
}
