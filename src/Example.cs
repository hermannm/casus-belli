using System;
using Godot;

namespace Immerse.BfhClient;

public partial class Example : Sprite2D
{
	private double _timePassed;
	[Export] private double _amplitude = 10.0;

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
