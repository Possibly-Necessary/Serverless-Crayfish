package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Define a struct for testing the objective function of the benchmark
type ObjectiveFunction struct {
	dim      int                         // Problem/benchmark dimension -- variable, I decide what it is
	Lb       float64                     // Lowerbound lb
	Ub       float64                     // Upperbound ub
	Goal     float64                     // optimization goal (error threshold)
	Evaluate func(vec []float64) float64 // Objective function
}

// Create a variable of the struct for the Schwefel fucntion
var Sch = ObjectiveFunction{
	dim:      30,   //change these for the Schwefel
	Lb:       -500, // Lower & Upper-bound associated with the Schwefel function
	Ub:       500,
	Goal:     1e-5,
	Evaluate: Schwefel,
}

// https://www.sfu.ca/~ssurjano/optimization.html
// https://www.sfu.ca/~ssurjano/schwef.html
// Benchmark function: Schwefel Fucntion
func Schwefel(vec []float64) float64 {
	sum := 0.0
	for _, xi := range vec { //figure this out later
		sum += xi * math.Sin(math.Sqrt(math.Abs(xi)))
	}
	return 418.9829*float64(len(vec)) - sum
}

//rand.Seed(time.Now().UnixNano())

func initializePopulation(N, dim int, lb, ub []float64) [][]float64 {
	X := make([][]float64, N) // Population matrix

	// Check if there is only one boundary for all dimensions -- from Matlab "if B_no==1"
	singleBound := (len(lb) == 1 && len(ub) == 1)

	for i := range X {
		X[i] = make([]float64, dim)
		for j := range X[i] {
			if singleBound {
				// Use the single boundary for all dimensions
				X[i][j] = lb[0] + rand.Float64()*(ub[0]-lb[0])
			} else {
				// Use the respective boundary for each dimension
				X[i][j] = lb[j] + rand.Float64()*(ub[j]-lb[j])
			}
		}
	}
	return X //return the matrix
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	// Test test... Population
	N := 10
	dim := 5
	// Boundaries (upper-lower) for solution space -- boundries depend on the specific banchmark/problem fucntion used
	lb := []float64{-5} // single lower bound of -5 for all dimensions
	ub := []float64{5}  // single upper bound of 5 for all dimensions
	population := initializePopulation(N, dim, lb, ub)

	fmt.Println(population)

	// // Testing Schwefel stuff
	// x := []float64{420.9687, 420.9687} //2D input

	// result := Schwefel(x)
	// println("Schwefel function result:", result)

}
