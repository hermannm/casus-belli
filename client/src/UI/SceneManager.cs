using System;
using Godot;

namespace CasusBelli.Client.UI;

public partial class SceneManager : Node
{
    /// SceneManager singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static SceneManager Instance { get; private set; } = null!;

    private string _currentScenePath = ScenePaths.MainMenu;
    private string? _previousScenePath = null;

    public override void _EnterTree()
    {
        // ReSharper disable once ConditionIsAlwaysTrueOrFalseAccordingToNullableAPIContract
        if (Instance is null)
        {
            Instance = this;
        }
    }

    public void LoadScene(string scenePath)
    {
        var err = GetTree().ChangeSceneToFile(scenePath);
        if (err != Error.Ok)
        {
            MessageDisplay.Instance.ShowError("Failed to load scene", err.ToString());
            return;
        }

        _previousScenePath = _currentScenePath;
        _currentScenePath = scenePath;
    }

    public void LoadPreviousScene()
    {
        if (_previousScenePath != null)
        {
            LoadScene(_previousScenePath);
        }
    }
}
