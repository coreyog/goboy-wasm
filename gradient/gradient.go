package gradient

import (
	colorful "github.com/lucasb-eyer/go-colorful"
)

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

var Keypoints GradientTable

func init() {
	Keypoints = GradientTable{
		{MustParseHex("#FF0000"), 0.0},
		{MustParseHex("#00FF00"), 0.33},
		{MustParseHex("#0000FF"), 0.66},
		{MustParseHex("#FF0000"), 1.0},
	}
}

func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}

func (self GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(self)-1; i++ {
		c1 := self[i]
		c2 := self[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return self[len(self)-1].Col
}
