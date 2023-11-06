using Godot;
using Immerse.BfhClient.UI;

namespace Immerse.BfhClient;

public partial class SceneManager : Node
{
    /// SceneManager singleton instance.
    /// Should never be null, since it is configured to autoload in Godot, and set in _EnterTree.
    public static SceneManager Instance { get; private set; } = null!;

    private string _currentScene = Scenes.MainMenu;
    private string _previousScene = "";

    public override void _EnterTree()
    {
        // ReSharper disable once ConditionIsAlwaysTrueOrFalseAccordingToNullableAPIContract
        if (Instance is null)
        {
            Instance = this;
        }
    }

    public void LoadScene(string scene)
    {
        var err = GetTree().ChangeSceneToFile(scene);
        if (err != Error.Ok)
        {
            MessageDisplay.Instance.ShowError("Failed to load scene", err.ToString());
            return;
        }

        _previousScene = _currentScene;
        _currentScene = scene;
    }

    public void LoadPreviousScene()
    {
        LoadScene(_previousScene);
    }
}
