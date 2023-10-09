using System;
using Godot;
using Immerse.BfhClient.Api;

namespace Immerse.BfhClient.Menus.ServerAddressMenu;

public partial class ConnectButton : Button
{
    private ApiClient _apiClient = null!;
    private TextEdit _serverAddressField = null!;

    public override void _Ready()
    {
        _apiClient = this.GetApiClient();
        _serverAddressField = GetNode<TextEdit>("ServerAddressField");
    }

    public override void _Pressed()
    {
        try
        {
            _apiClient.Connect(_serverAddressField.Text);
        }
        catch (Exception e)
        {
            GD.PushError(e.Message);
            throw;
        }
    }
}
