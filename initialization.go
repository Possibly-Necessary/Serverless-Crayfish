package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Define a struct for testing the objective function/benchmark
type ObjectiveFunction struct {
	dim int     // Problem/benchmark dimension -- variable, I decide what it is
	Lb  float64 // Lowerbound lb
	Ub  float64 // Upperbound ub
	//Goal     float64                     // optimization goal (error threshold)
	Evaluate func(vec []float64) float64 // Objective function
}

// Create a variable of type struct for the Schwefel fucntion (or it could be any other function)
var ObjF = ObjectiveFunction{
	dim: 10,   // change these for the Schwefel
	Lb:  -500, // Lower & Upper-bound associated with the Schwefel function (-500, 500)
	Ub:  500,
	//Goal:     1e-5,
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

// Equation 4
func p_obj(x float64) float64 {
	return 0.2 * (1 / (math.Sqrt(2*math.Pi) * 3)) * math.Exp(-math.Pow(x-25, 2)/(2*math.Pow(3, 2)))
}

//func CrayFish(N, T, dim int, lb, ub float64, ObjF func([]float64)) (float64, []float64, []float64, []float64){

//}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	//T := 100 // Max iteration
	//N := 10  // Population
	//dim := 5
	// Boundaries (upper-lower) for solution space -- boundries depend on the specific banchmark/problem fucntion used
	//lb := []float64{-5} // single lower bound of -5 for all dimensions
	//ub := []float64{5}  // single upper bound of 5 for all dimensions

	//fmt.Println(population)

	// // Testing Schwefel stuff
	// x := []float64{420.9687, 420.9687} //2D input

	// result := Schwefel(x)
	// println("Schwefel function result:", result)

	N := 10
	T := 300 //iteration

	// Getting the variables for the population intialization
	dim := ObjF.dim

	lb := make([]float64, dim)
	ub := make([]float64, dim)

	populationX := initializePopulation(N, dim, lb, ub)

	// Parameters
	var (
		cuF         []float64 = make([]float64, T)
		globalCov   []float64 = make([]float64, T) // zero row vector of size T
		BestFitness           = math.Inf(1)
		BestPos     []float64 = make([]float64, dim)
		fitnessF    []float64 = make([]float64, N)
		GlobalPos   []float64 = make([]float64, dim)
	)
	// Calculating the fitness value of the function
	for i := 0; i < N; i++ {
		fitnessF[i] = ObjF.Evaluate((populationX[i])) // X[i] is a slice representing the i-th row in X
		if fitnessF[i] < BestFitness {
			BestFitness = fitnessF[i]
			copy(BestPos, populationX[i]) // Make a copy of population vector to avoid referencing issues
		}
	}

	copy(GlobalPos, BestPos)
	GlobalFitness := BestFitness
	cuF[0] = BestFitness // Assign value to the first element of this vector

	t := 1
	for t < T {
		//Decreasing curve --> Equation 7
		C := 2 - (float64(t) / float64(T))
		//Define the temprature from Equation 3
		temp := rand.Float64()*15 + 20

		Xf := make([]float64, dim)
		for i := 0; i < dim; i++ {
			Xf[i] = (BestPos[i] + GlobalPos[i]) / 2
		}

		Xfood := make([]float64, dim)
		copy(Xfood, BestPos) // copy elemnts from one slice (vector) into another
		Xnew := make([][]float64, N)

		for i := 0; i < N; i++ {
			Xnew[i] = make([]float64, dim)
			if temp > 30 { // Summer resort stage
				if rand.Float64() < 0.5 {
					for j := 0; j < dim; j++ { // Equation 6
						Xnew[i][j] = populationX[i][j] + C*rand.Float64()*(Xf[j]-populationX[i][j])
					}
				} else { // Competition Stage
					for j := 0; j < dim; j++ {
						z := rand.Intn(N) // Random crayfish --> Here removed 20
						//z := math.Round(rand.Float64()*(N-1)) + 1 --> or try this
						Xnew[i][j] = populationX[i][j] - populationX[z][j] + Xf[j]
					}
				}
			} else { //Foraging stage
				P := 3 * rand.Float64() * fitnessF[i] / ObjF.Evaluate(Xfood)
				if P > 2 {
					//Food is broken down becuase it's too big
					for j := 0; j < dim; j++ {
						Xfood[j] *= math.Exp(-1 / P)
						Xnew[i][j] = populationX[i][j] + math.Cos(2*math.Pi*rand.Float64())*Xfood[j]*p_obj(temp) - math.Sin(2*math.Pi*rand.Float64())*Xfood[j]*p_obj(temp)
					}
				} else {
					for j := 0; j < dim; j++ {
						Xnew[i][j] = (populationX[i][j] - Xfood[j]) * p_obj(temp) * rand.Float64() * populationX[i][j]
					}
				}
			}
		}

		// Boundary conditions
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

		//Global stuff
		copy(GlobalPos, Xnew[0])
		GlobalFitness = ObjF.Evaluate(GlobalPos)

		for i := 0; i < N; i++ {
			NewFitness := ObjF.Evaluate(Xnew[i])
			if NewFitness < GlobalFitness {
				GlobalFitness = NewFitness
				copy(GlobalPos, Xnew[i])
			}

			// Update population to a new location

			if NewFitness < fitnessF[i] {
				fitnessF[i] = NewFitness
				copy(populationX[i], Xnew[i])
				if fitnessF[i] < BestFitness {
					copy(BestPos, populationX[i])
				}
			}
		}

		globalCov[t] = GlobalFitness
		//Or try this:
		//globalCov = append(globalCov, GlobalFitness)
		cuF[t] = BestFitness
		//cuF = append(cuF, BestFitness)

		t++

		if t%50 == 0 { //Print the best fitness every 50 interation (after it passes 100 iteration)
			fmt.Printf("COA iter %d: %f\n", t, BestFitness)
		}

	}

	fmt.Println(BestFitness)
	//For the function:
	//Return BestFitness, BestPos, cuF, Globalcov

}
