using Godot;

namespace CasusBelli.Client.UI;

public partial class BackButton : Button
{
    public override void _Pressed()
    {
        SceneManager.Instance.LoadPreviousScene();
    }
}
