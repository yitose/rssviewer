package db

import (
	"fmt"
	"testing"

	"github.com/yitose/rssviewer/internal/color"
)

func print(colorcode int) {
	fmt.Printf("\033[38;5;%dm%03d\033[0m ", colorcode, colorcode)
}

func TestColor(t *testing.T) {
	colors := color.GetColorRange(
		defaultMaxHue,
		defaultMinHue,
		defaultMaxSaturatio,
		defaultMinSaturatio,
		defaultMaxLightness,
		defaultMinLightness,
	)
	for _, i := range colors {
		print(i)
	}
	fmt.Printf("\n%d colors\n", len(colors))
}
