using System;
using System.Collections.Generic;
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
    private readonly HttpClient _httpClient = new();
    private readonly ClientWebSocket _websocket = new();
    private readonly CancellationTokenSource _cancellation = new();
    private readonly MessageSender _messageSender;
    private readonly MessageReceiver _messageReceiver;
    private bool _lobbyJoined = false;

    public ApiClient()
    {
        _messageSender = new MessageSender(_websocket);
        _messageReceiver = new MessageReceiver(_websocket);
    }

    public override void _EnterTree()
    {
        Instance = this;
        AddMessageReceivedSignals();
        this.AddServerMessageHandler<ErrorMessage>(DisplayServerError);

        if (OS.IsDebugBuild())
            TryConnect("localhost:8000");
    }

    public override void _Process(double delta)
    {
        if (!_lobbyJoined)
            return;

        if (!_messageReceiver.MessageQueue.TryDequeue(out var message))
            return;

        var signal = GetMessageReceivedSignalName(message.Tag);
        var err = EmitSignal(signal, message.Data);
        if (err != Error.Ok)
            GD.PushError($"Failed to emit signal '{signal}': {err}");
    }

    private void AddMessageReceivedSignals()
    {
        foreach (var tag in MessageDictionary.ReceivableMessageTags.Keys)
        {
            // https://cyoann.github.io/GodotSharpAPI/html/9ad4493d-f079-22f2-cdaf-37cb2660f34b.htm
            var signalArg = new GodotDictionary();
            // ReSharper disable once BuiltInTypeReferenceStyle
            signalArg.Add("message", (Int32)Variant.Type.Object);
            var signalArgs = new GodotArray();
            signalArgs.Add(signalArg);
            AddUserSignal(GetMessageReceivedSignalName(tag), signalArgs);
        }
    }

    public static string GetMessageReceivedSignalName(MessageTag tag)
    {
        return tag + "MessageReceived";
    }

    public bool TryConnect(string serverUrl)
    {
        Uri parsedUrl;
        try
        {
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
        return _lobbyJoined ? LeaveLobby() : Task.CompletedTask;
    }

    /// <summary>
    /// Sends the given message to the server.
    /// </summary>
    ///
    /// <typeparam name="TMessage">
    /// Must be registered in <see cref="RegisterSendableMessages"/>, which should be all message
    /// types marked with <see cref="ISendableMessage"/>.
    /// </typeparam>
    public void SendServerMessage<TMessage>(TMessage message)
        where TMessage : GodotObject, ISendableMessage
    {
        _messageSender.SendQueue.Add(message);
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
    public async Task<bool> TryJoinLobby(string lobbyName, string username)
    {
        if (ServerUrl is null)
        {
            MessageDisplay.Instance.ShowError("Tried to join lobby before setting server URL");
            return false;
        }

        new Thread(() => _messageReceiver.ReadMessagesIntoQueue(_cancellation.Token)).Start();
        new Thread(() => _messageSender.SendMessagesFromQueue(_cancellation.Token)).Start();

        var joinLobbyUrl = new UriBuilder(ServerUrl)
        {
            Path = "join",
            Query = $"lobbyName={lobbyName}&username={username}"
        };

        try
        {
            await _websocket.ConnectAsync(joinLobbyUrl.Uri, _cancellation.Token);
        }
        catch (Exception e)
        {
            _cancellation.Cancel();
            MessageDisplay.Instance.ShowError(
                "Failed to create WebSocket connection to server",
                e.Message
            );
            return false;
        }

        _lobbyJoined = true;
        return true;
    }

    public Task LeaveLobby()
    {
        _cancellation.Cancel();
        _lobbyJoined = false;

        return _websocket.CloseAsync(
            WebSocketCloseStatus.NormalClosure,
            "Client initiated disconnect from game server",
            _cancellation.Token
        );
    }

    private static void DisplayServerError(ErrorMessage errorMessage)
    {
        MessageDisplay.Instance.ShowError(errorMessage.Error);
    }
}
