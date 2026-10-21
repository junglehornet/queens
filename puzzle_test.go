package main

import (
	"math/rand"
	"testing"
)

func TestGeneratedPuzzles(t *testing.T) {
	for n := 5; n <= 9; n++ {
		for seed := int64(0); seed < 8; seed++ {
			p := generate(n, rand.New(rand.NewSource(seed)))
			if p.solutions(2) != 1 {
				t.Fatalf("size %d seed %d is not unique", n, seed)
			}
			for region := 0; region < n; region++ {
				start := -1
				total := 0
				for i, v := range p.regions {
					if v == region {
						start = i
						total++
					}
				}
				if start < 0 {
					t.Fatal("missing region")
				}
				seen := map[int]bool{start: true}
				queue := []int{start}
				for len(queue) > 0 {
					i := queue[0]
					queue = queue[1:]
					for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
						r, c := i/n+d[0], i%n+d[1]
						j := r*n + c
						if r >= 0 && r < n && c >= 0 && c < n && !seen[j] && p.regions[j] == region {
							seen[j] = true
							queue = append(queue, j)
						}
					}
				}
				if len(seen) != total {
					t.Fatal("disconnected region")
				}
			}
		}
	}
}
