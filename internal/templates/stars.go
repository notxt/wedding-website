package templates

import (
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"strings"
)

const starSeed = 91827

func renderStars() template.HTML {
	r := rand.New(rand.NewSource(starSeed))
	var b strings.Builder

	for i := 0; i < 280; i++ {
		x := r.Float64() * 1400
		y := math.Pow(r.Float64(), 2.4) * 720
		var radius float64
		switch p := r.Float64(); {
		case p < 0.06:
			radius = 2.4
		case p < 0.22:
			radius = 1.6
		case p < 0.55:
			radius = 1.1
		default:
			radius = 0.6
		}
		opacity := (0.25 + r.Float64()*0.6) * 0.65
		fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="%.2f" fill="#e6d5b3" opacity="%.3f"/>`,
			x, y, radius, opacity)
	}

	clusters := []struct {
		cx, cy, spread float64
	}{
		{180, 100, 70}, {380, 180, 90}, {720, 70, 80}, {1080, 130, 100},
		{560, 240, 90}, {1240, 200, 90}, {220, 320, 80}, {960, 360, 80},
	}
	for _, c := range clusters {
		n := 14 + r.Intn(10)
		for i := 0; i < n; i++ {
			dx := (r.Float64() + r.Float64() + r.Float64() - 1.5) * c.spread
			dy := (r.Float64() + r.Float64() + r.Float64() - 1.5) * c.spread * 0.7
			var radius float64
			switch p := r.Float64(); {
			case p < 0.1:
				radius = 2.0
			case p < 0.4:
				radius = 1.3
			default:
				radius = 0.8
			}
			opacity := (0.35 + r.Float64()*0.55) * 0.65
			fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="%.2f" fill="#e6d5b3" opacity="%.3f"/>`,
				c.cx+dx, c.cy+dy, radius, opacity)
		}
	}

	features := []struct {
		x, y, r float64
	}{
		{240, 180, 2.4}, {820, 130, 2.2}, {1080, 460, 2.0},
		{560, 380, 1.8}, {380, 620, 2.2}, {1240, 360, 1.8},
	}
	for _, f := range features {
		fmt.Fprintf(&b,
			`<g><circle cx="%.0f" cy="%.0f" r="%.2f" fill="#e6d5b3" opacity="0.05"/><circle cx="%.0f" cy="%.0f" r="%.2f" fill="#e6d5b3" opacity="0.85"/></g>`,
			f.x, f.y, f.r*2.4, f.x, f.y, f.r)
	}

	return template.HTML(b.String())
}
