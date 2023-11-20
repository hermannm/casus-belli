using Godot;

namespace Immerse.BfhClient.UI;

public partial class BackButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
    }
}
