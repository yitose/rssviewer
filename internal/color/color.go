package color

import (
	"math"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

func getTcellColor(code int) tcell.Color {
	return tcell.Color(1<<32 + code)
}

func rgb2hsl(r int32, g int32, b int32) (int, int, int) {
	var rf, gf, bf, max, min, l, d, s, h float64

	rf = math.Max(math.Min(float64(r)/255, 1), 0)
	gf = math.Max(math.Min(float64(g)/255, 1), 0)
	bf = math.Max(math.Min(float64(b)/255, 1), 0)
	max = math.Max(rf, math.Max(gf, bf))
	min = math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2

	if max != min {
		d = max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}
		if max == rf {
			if gf < bf {
				h = (gf-bf)/d + 6
			} else {
				h = (gf - bf) / d
			}
		} else if max == gf {
			h = (bf-rf)/d + 2
		} else {
			h = (rf-gf)/d + 4
		}
	} else {
		h = 0
		s = 0
	}

	return int(h * 60), int(s * 100), int(l * 100)
}

func GetColorRange(maxHue, minHue, maxSaturatio, minSaturatio, maxLightness, minLightness int) []int {
	result := []int{}
	for i := 0; i < 256; i++ {
		h, s, l := rgb2hsl(getTcellColor(i).RGB())
		if maxHue >= h && h >= minHue &&
			maxSaturatio >= s && s >= minSaturatio &&
			maxLightness >= l && l >= minLightness {
			result = append(result, i)
		}
	}
	return result
}

func GetRandomColor(maxHue, minHue, maxSaturatio, minSaturatio, maxLightness, minLightness int) int {
	colors := GetColorRange(maxHue, minHue, maxSaturatio, minSaturatio, maxLightness, minLightness)
	return colors[rand.Intn(len(colors))]
}
