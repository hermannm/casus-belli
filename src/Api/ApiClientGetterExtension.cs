using System;
using Godot;

namespace Immerse.BfhClient.Api;

/// <summary>
/// Utility extension method for getting the global ApiClient instance from any node.
/// The ApiClient should always be available, since it is configured to autoload in Godot.
/// </summary>
public static class ApiClientGetterExtension
{
    public static ApiClient GetApiClient(this Node node)
    {
        return node.GetNode<ApiClient>("/root/ApiClient")
            ?? throw new Exception(
                "Failed to find ApiClient node - is it added in the project autoload list?"
            );
    }
}
