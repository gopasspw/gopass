package smetrics

import (
	"math"
)

func Ukkonen(a, b string, icost, dcost, scost int) int {
	var lowerCost int

	if icost < dcost && icost < scost {
		lowerCost = icost
	} else if dcost < scost {
		lowerCost = dcost
	} else {
		lowerCost = scost
	}

	infinite := math.MaxInt32 / 2

	var r []int
	var k, kprime, p, t int
	var ins, del, sub int

	if len(a) > len(b) {
		t = (len(a) - len(b) + 1) * lowerCost
	} else {
		t = (len(b) - len(a) + 1) * lowerCost
	}

	for {
		if (t / lowerCost) < (len(b) - len(a)) {
			continue
		}

		// This is the right damn thing since the original Ukkonen
		// paper minimizes the expression result only, but the uncommented version
		// doesn't need to deal with floats so it's faster.
		// p = int(math.Floor(0.5*((float64(t)/float64(lowerCost)) - float64(len(b) - len(a)))))
		p = ((t / lowerCost) - (len(b) - len(a))) / 2

		k = -p
		kprime = k

		rowlength := (len(b) - len(a)) + (2 * p)

		r = make([]int, rowlength+2)

		for i := 0; i < rowlength+2; i++ {
			r[i] = infinite
		}

		for i := 0; i <= len(a); i++ {
			for j := 0; j <= rowlength; j++ {
				if i == j+k && i == 0 {
					r[j] = 0
				} else {
					if j-1 < 0 {
						ins = infinite
					} else {
						ins = r[j-1] + icost
					}

					del = r[j+1] + dcost
					sub = r[j] + scost

					if i-1 < 0 || i-1 >= len(a) || j+k-1 >= len(b) || j+k-1 < 0 {
						sub = infinite
					} else if a[i-1] == b[j+k-1] {
						sub = r[j]
					}

					if ins < del && ins < sub {
						r[j] = ins
					} else if del < sub {
						r[j] = del
					} else {
						r[j] = sub
					}
				}
			}
			k++
		}

		if r[(len(b)-len(a))+(2*p)+kprime] <= t {
			break
		} else {
			t *= 2
		}
	}

	return r[(len(b)-len(a))+(2*p)+kprime]
}
