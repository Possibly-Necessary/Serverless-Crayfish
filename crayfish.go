/*

 Main reference of this and the original code is here: https://github.com/rao12138/COA-s-code/tree/main/COA
 Source: https://dl.acm.org/doi/10.1007/s10462-023-10567-4

 Complexity of the Crayfish Optimization Algorithm (COA):
	O(COA) = O(algorithm parameter definition) + O(population initialization) + O(function evaluation cost) + O(population update)
		   = O(1) + O(N*D) + O(T*N*C) + O(T*N*D)
		   = O(T*N*(C+D))
	where N is the number of crayfish, D is the dimension of the problem, T is the number of iteration of COA (usually a large value),
	and C is the evaluation cost of the function.

*/

package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// https://www.sfu.ca/~ssurjano/optimization.html
// Benchmark function: this is the F8 function in the original paper, which seems to be a variant of Schwefel
func Schwefel(vec []float64) float64 {
	sum := 0.0
	for _, xi := range vec {
		//sum += (-xi * math.Sin(math.Sqrt(math.Abs(xi))))
		sum += (-xi * math.Sin(math.Sqrt(math.Abs(xi))))
	}
	//return 418.9829*float64(len(vec)) - sum
	return sum
}

// Equation 4: Mathimatical model of crayfish intake
func p_obj(x float64) float64 {
	return 0.2 * (1 / (math.Sqrt(2*math.Pi) * 3)) * math.Exp(-math.Pow(x-25, 2)/(2*math.Pow(3, 2)))
}

// Function to initialize the population
func initializePopulation(N, dim int, lb, ub float64) [][]float64 {

	// Initialize the population N x Dim matrix, X
	X := make([][]float64, N)
	for i := 0; i < N; i++ {
		X[i] = make([]float64, dim)
	}

	for i := range X {
		for j := range X[i] {
			X[i][j] = rand.Float64()*(ub-lb) + lb
		}
	}
	return X
}

func main() {

	rand.Seed(time.Now().UnixNano())

	N := 30      // Population
	dim := 10    // Dimension
	LB := -500.0 // Lower bound (associated with the benchmark function)
	UB := 500.0  // Upper bound
	T := 500     // Maximum iterations

	lb := make([]float64, 1)
	ub := make([]float64, 1)

	lb[0] = -500.0
	ub[0] = 500.0

	// Initialize the arrays
	var (
		cuF         []float64 = make([]float64, T)
		globalCov   []float64 = make([]float64, T) // zero row vector of size T
		BestFitness           = math.Inf(1)
		BestPos     []float64 = make([]float64, dim)
		fitnessF    []float64 = make([]float64, N)
		GlobalPos   []float64 = make([]float64, dim)
	)

	X := initializePopulation(N, dim, LB, UB) // Generate the population matrix

	for i := 0; i < N; i++ {
		fitnessF[i] = Schwefel(X[i]) // Get the fitness value from the benchmark function
		if fitnessF[i] < BestFitness {
			BestFitness = fitnessF[i]
			copy(BestPos, X[i])
		}
	}

	// Update best position to Global position
	copy(GlobalPos, BestPos)
	GlobalFitness := BestFitness
	cuF[0] = BestFitness

	Xf := make([]float64, dim) // For Xshade -- array for the cave
	Xfood := make([]float64, dim)

	Xnew := make([][]float64, N) // Initializing a 2d array
	for i := 0; i < N; i++ {
		Xnew[i] = make([]float64, dim)
	}

	// Start the timer
	kickStart := time.Now()
	t := 0
	for t < T {
		//Decreasing curve --> Equation 7
		C := 2 - (float64(t) / float64(T))
		//Define the temprature from Equation 3
		tmp := rand.Float64()*15 + 20

		for i := 0; i < dim; i++ { // Calculating the Cave -> Xshade = XL + XG/2
			Xf[i] = (BestPos[i] + GlobalPos[i]) / 2
		}
		copy(Xfood, BestPos) // copy the best position to the Xfood vector

		for i := 0; i < N; i++ {
			//Xnew[i] = make([]float64, dim) //--> took this part out
			if tmp > 30 { // Summer resort stage
				if rand.Float64() < 0.5 {
					for j := 0; j < dim; j++ { // Equation 6
						Xnew[i][j] = X[i][j] + C*rand.Float64()*(Xf[j]-X[i][j])
					}
				} else { // Competition Stage
					for j := 0; j < dim; j++ {
						z := rand.Intn(N) // Random crayfish
						//z := math.Round(rand.Float64()*(N-1)) + 1 //--> or try this
						Xnew[i][j] = X[i][j] - X[z][j] + Xf[j] // Equation 8
					}
				}
			} else { // Foraging stage
				P := 3 * rand.Float64() * fitnessF[i] / Schwefel(Xfood)
				if P > 2 {
					//Food is broken down becuase it's too big
					for j := 0; j < dim; j++ {
						Xfood[j] *= math.Exp(-1 / P)
						Xnew[i][j] = X[i][j] + math.Cos(2*math.Pi*rand.Float64())*Xfood[j]*p_obj(tmp) - math.Sin(2*math.Pi*rand.Float64())*Xfood[j]*p_obj(tmp)
					} // ^^ Equation 13: crayfish foraging
				} else {
					for j := 0; j < dim; j++ { // The case where the food is a moderate size
						Xnew[i][j] = (X[i][j]-Xfood[j])*p_obj(tmp) + p_obj(tmp)*rand.Float64()*X[i][j]
					}
				}
			}
		}

		// Boundary conditions checks
		for i := 0; i < N; i++ {
			for j := 0; j < dim; j++ {
				if len(ub) == 1 {
					Xnew[i][j] = math.Min(ub[0], Xnew[i][j])
					Xnew[i][j] = math.Max(lb[0], Xnew[i][j])
				} else {
					Xnew[i][j] = math.Min(ub[j], Xnew[i][j])
					Xnew[i][j] = math.Max(lb[j], Xnew[i][j])
				}
			}
		}

		//Global update stuff
		copy(GlobalPos, Xnew[0])
		GlobalFitness = Schwefel(GlobalPos)

		for i := 0; i < N; i++ {
			NewFitness := Schwefel(Xnew[i])
			if NewFitness < GlobalFitness {
				GlobalFitness = NewFitness
				copy(GlobalPos, Xnew[i])
			}

			// Update population to a new location
			if NewFitness < fitnessF[i] {
				fitnessF[i] = NewFitness
				copy(X[i], Xnew[i])
				if fitnessF[i] < BestFitness {
					BestFitness = fitnessF[i]
					copy(BestPos, X[i])
				}
			}
		}

		globalCov[t] = GlobalFitness
		cuF[t] = BestFitness

		t++

		if t%50 == 0 { //Print the best fitness every 50 interation (after it passes 100 iteration)
			fmt.Printf("COA iteration %d: %f\n", t, BestFitness)
		}
	}

	fmt.Println("Optimal value of the objective function found by COA:\n", BestFitness)
	fmt.Println("Best solution obtained by COA: \n", BestPos)
	fmt.Println("Executed in:\n", time.Since(kickStart))
}
