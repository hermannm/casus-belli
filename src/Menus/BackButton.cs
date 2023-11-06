using Godot;

namespace Immerse.BfhClient.Menus;

public partial class BackButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
    }
}
