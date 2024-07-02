package plot

import (
	"fmt"
	"image/color"
)

func RGBString(r, g, b uint8) string {
	return fmt.Sprintf(`"rgb(%d,%d,%d,0)"`, r, g, b)
}

type WeightedColor struct {
	Value float64
	Color color.RGBA
}

func (c WeightedColor) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`[%f,"rgb(%d,%d,%d,%f)"]`,
		c.Value, c.Color.R, c.Color.G, c.Color.B, float64(c.Color.A)/255)
	return []byte(str), nil
}

// https://mdigi.tools/color-shades/
var BlueShades = []WeightedColor{
	{Value: 0.0, Color: color.RGBA{0xea, 0xf8, 0xfd, 1}},
	{Value: 0.1, Color: color.RGBA{0xbf, 0xeb, 0xfa, 1}},
	{Value: 0.2, Color: color.RGBA{0x94, 0xdd, 0xf6, 1}},
	{Value: 0.3, Color: color.RGBA{0x69, 0xd0, 0xf2, 1}},
	{Value: 0.4, Color: color.RGBA{0x3f, 0xc2, 0xef, 1}},
	{Value: 0.5, Color: color.RGBA{0x14, 0xb5, 0xeb, 1}},
	{Value: 0.6, Color: color.RGBA{0x10, 0x94, 0xc0, 1}},
	{Value: 0.7, Color: color.RGBA{0x0d, 0x73, 0x96, 1}},
	{Value: 0.8, Color: color.RGBA{0x09, 0x52, 0x6b, 1}},
	{Value: 0.9, Color: color.RGBA{0x05, 0x31, 0x40, 1}},
	{Value: 1.0, Color: color.RGBA{0x02, 0x10, 0x15, 1}},
}

var PinkShades = []WeightedColor{
	{Value: 0.0, Color: color.RGBA{0xfe, 0xe7, 0xf3, 1}},
	{Value: 0.1, Color: color.RGBA{0xfc, 0xb6, 0xdc, 1}},
	{Value: 0.2, Color: color.RGBA{0xf9, 0x85, 0xc5, 1}},
	{Value: 0.3, Color: color.RGBA{0xf7, 0x55, 0xae, 1}},
	{Value: 0.4, Color: color.RGBA{0xf5, 0x24, 0x96, 1}},
	{Value: 0.5, Color: color.RGBA{0xdb, 0x0a, 0x7d, 1}},
	{Value: 0.6, Color: color.RGBA{0xaa, 0x08, 0x61, 1}},
	{Value: 0.7, Color: color.RGBA{0x7a, 0x06, 0x45, 1}},
	{Value: 0.8, Color: color.RGBA{0x49, 0x03, 0x2a, 1}},
	{Value: 0.9, Color: color.RGBA{0x18, 0x01, 0x0e, 1}},
	{Value: 1.0, Color: color.RGBA{0x00, 0x00, 0x00, 1}},
}

var GreenShades = []WeightedColor{
	{Value: 0.0, Color: color.RGBA{0xed, 0xf7, 0xf2, 0}},
	{Value: 0.1, Color: color.RGBA{0xc9, 0xe8, 0xd7, 0}},
	{Value: 0.2, Color: color.RGBA{0xa5, 0xd9, 0xbc, 0}},
	{Value: 0.3, Color: color.RGBA{0x81, 0xca, 0xa2, 0}},
	{Value: 0.4, Color: color.RGBA{0x5e, 0xbb, 0x87, 0}},
	{Value: 0.5, Color: color.RGBA{0x44, 0xa1, 0x6e, 0}},
	{Value: 0.6, Color: color.RGBA{0x35, 0x7e, 0x55, 0}},
	{Value: 0.7, Color: color.RGBA{0x26, 0x5a, 0x3d, 0}},
	{Value: 0.8, Color: color.RGBA{0x17, 0x36, 0x25, 0}},
	{Value: 0.9, Color: color.RGBA{0x08, 0x12, 0x0c, 0}},
	{Value: 1.0, Color: color.RGBA{0x00, 0x00, 0x00, 0}},
}
