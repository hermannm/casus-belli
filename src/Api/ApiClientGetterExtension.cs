using System;
using Godot;

namespace Immerse.BfhClient.Api;

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
