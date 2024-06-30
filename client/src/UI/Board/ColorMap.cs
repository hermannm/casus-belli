using System;
using System.Collections.Generic;
using System.Text.Json;
using Godot;

namespace CasusBelli.Client.UI.Board;

public class ColorMap
{
    private readonly Dictionary<string, string> _map;

    public ColorMap(string boardId)
    {
        var path = $"res://assets/board/{boardId}/colormap.json";
        var file = FileAccess.Open(path, FileAccess.ModeFlags.Read);
        if (file == null)
        {
            throw new Exception($"Failed to find color map at path '{path}'");
        }

        var map = JsonSerializer.Deserialize<Dictionary<string, string>>(file.GetAsText());
        if (map == null)
        {
            throw new Exception($"Failed to deserialize colormap at path '{path}'");
        }

        _map = map;
    }

    public string? GetRegionNameForColor(string color)
    {
        if (_map.TryGetValue(color, out var regionName))
        {
            return regionName;
        }
        else
        {
            return null;
        }
    }
}
