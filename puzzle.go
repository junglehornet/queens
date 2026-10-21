package main

import "math/rand"

type puzzle struct {
	n       int
	regions []int
}

// Generate connected regions around a legal queen permutation, then reject
// boards with multiple solutions. The solver never relies on the seed queens.
func generate(n int, rng *rand.Rand) puzzle {
	for {
		queens := rng.Perm(n)
		valid := true
		for r := 1; r < n; r++ {
			if abs(queens[r]-queens[r-1]) <= 1 {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		p := puzzle{n: n, regions: make([]int, n*n)}
		for i := range p.regions {
			p.regions[i] = -1
		}
		for r, c := range queens {
			p.regions[r*n+c] = r
		}
		for remaining := n*n - n; remaining > 0; remaining-- {
			type edge struct{ cell, region int }
			var frontier []edge
			for i, region := range p.regions {
				if region < 0 {
					continue
				}
				for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					r, c := i/n+d[0], i%n+d[1]
					if r >= 0 && r < n && c >= 0 && c < n && p.regions[r*n+c] < 0 {
						frontier = append(frontier, edge{r*n + c, region})
					}
				}
			}
			e := frontier[rng.Intn(len(frontier))]
			p.regions[e.cell] = e.region
		}
		if p.solutions(2) == 1 {
			return p
		}
	}
}

func (p puzzle) solutions(limit int) int {
	count := 0
	var visit func(int, uint, uint, int)
	visit = func(row int, cols, regions uint, prev int) {
		if count >= limit {
			return
		}
		if row == p.n {
			count++
			return
		}
		for c := 0; c < p.n; c++ {
			region := p.regions[row*p.n+c]
			if cols&(1<<c) != 0 || regions&(1<<region) != 0 || (row > 0 && abs(c-prev) <= 1) {
				continue
			}
			visit(row+1, cols|1<<c, regions|1<<region, c)
		}
	}
	visit(0, 0, 0, -2)
	return count
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
