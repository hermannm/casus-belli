using System;
using Godot;
using Immerse.BfhClient.Api;

namespace Immerse.BfhClient;

public partial class Example : Sprite2D
{
    [Export]
    private double _amplitude = 10.0;

    private double _timePassed;

    private ApiClient _apiClient = null!;

    public override void _Ready()
    {
        _apiClient = this.GetApiClient();
    }

    // Called every frame. 'delta' is the elapsed time since the previous frame.
    public override void _Process(double delta)
    {
        _timePassed += delta;

        Position = new Vector2(
            (float)(_amplitude + _amplitude * Math.Sin(_timePassed * 2.0)),
            (float)(_amplitude + _amplitude * Math.Cos(_timePassed * 1.5))
        );
    }
}
