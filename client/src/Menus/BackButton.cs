using Godot;

namespace CasusBelli.Client.Menus;

public partial class BackButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
    }
}
