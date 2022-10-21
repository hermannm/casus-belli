// Inspired by https://twin.sh/articles/35/how-to-add-colors-to-your-console-terminal-output-in-go
package color

import (
	"fmt"
	"runtime"
)

type Color string

var (
	Reset  Color = "\033[0m"
	Red    Color = "\033[31m"
	Green  Color = "\033[32m"
	Yellow Color = "\033[33m"
	Blue   Color = "\033[34m"
	Purple Color = "\033[35m"
	Cyan   Color = "\033[36m"
	Gray   Color = "\033[37m"
	White  Color = "\033[97m"
)

func (c Color) String(s string) string {
	return fmt.Sprintf("%s%s%s", c, s, Reset)
}

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}
}
