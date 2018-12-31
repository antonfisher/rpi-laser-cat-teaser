package detector

import (
	"fmt"
)

//FindCenter of binary presented shape
func FindCenter(a [][]int) (int, int) {
	var cx, cy, stepsPerformed int
	var w = len(a)
	var h = len(a[0])

	aCopy := make([][]int, w)
	for i := range aCopy {
		aCopy[i] = make([]int, h)
	}

	for {
		stepsPerformed++

		for x, row := range a {
			for y, v := range row {
				aCopy[x][y] = v
			}
		}

		var stepMax int
		for x, row := range a {
			for y := range row {
				neighbours := [][]int{
					[]int{x - 1, y - 1},
					[]int{x - 1, y},
					[]int{x - 1, y + 1},
					[]int{x, y + 1},
					[]int{x, y - 1},
					[]int{x + 1, y - 1},
					[]int{x + 1, y},
					[]int{x + 1, y + 1},
				}
				var sum int
				for _, n := range neighbours {
					xn := n[0]
					yn := n[1]
					if xn > 0 && yn > 0 && xn < w && yn < h {
						sum += aCopy[xn][yn]
					}
				}
				if sum < 8 {
					a[x][y] = 0
				}
				if sum > stepMax {
					stepMax = sum
					cx = x
					cy = y
				}
			}
		}
		if stepMax < 8 {
			break
		}
	}

	fmt.Printf("steps performed: %v\t", stepsPerformed)

	return cx, cy
}
