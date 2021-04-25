package pb_fuzz_workshop

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

type point struct {
	x int
	y int
}

type circle struct {
	center point
	radius int
}

type line struct {
	a point
	b point
}

type shape interface {
	Print()
}

func (c circle) Print() {
	fmt.Printf("center %v radius: %v\n", c.center, c.radius)
}

func (l line) Print() {
	fmt.Printf("a %v b: %v\n", l.a, l.b)
}

func GeneratePoint() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) point {
		return point{
			x: rapid.Int().Draw(t, "point_x").(int),
			y: rapid.Int().Draw(t, "point_y").(int),
		}
	})
}

func GenerateCircle() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) circle {
		return circle{
			center: point{
				x: rapid.Int().Draw(t, "circle_x").(int),
				y: rapid.Int().Draw(t, "circle_y").(int),
			},
			radius: rapid.Int().Draw(t, "rad").(int),
		}
	})
}

func GenerateLine() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) line {
		return line{
			a: point{
				x: rapid.Int().Draw(t, "line_x").(int),
				y: rapid.Int().Draw(t, "line_y").(int),
			},
			b: GeneratePoint().Draw(t, "line_b").(point),
		}
	})
}

func GenerateShape() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) shape {
		return rapid.OneOf(GenerateCircle(), GenerateLine()).Draw(t, "shape").(shape)
	})
}

func TestCustomStruct(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sp := rapid.SliceOf(GenerateShape()).Draw(t, "slice of points").([]shape)

		fmt.Println("-----------")
		for i := range sp {
			sp[i].Print()
		}
		fmt.Println("-----------")
	})
}
