// Crayfish Algorithm with dividing the population and using Go's concurrency 
package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

// Benchmark function
func F6(x []float64) float64 {
	var o float64
	for _, value := range x {
		o += math.Pow(math.Abs(value+0.5), 2)
	}
	return o
}

// Equation 4: Mathimatical model of crayfish intake
func p_obj(x float64) float64 {
	return 0.2 * (1 / (math.Sqrt(2*math.Pi) * 3)) * math.Exp(-math.Pow(x-25, 2)/(2*math.Pow(3, 2)))
}

// Function to initialize the (full) population
func initializePopulation(N, dim int, lb, ub []float64) [][]float64 {

	// Initialize the population N x Dim matrix, X
	X := make([][]float64, N)
	for i := 0; i < N; i++ {
		X[i] = make([]float64, dim)
	}

	for i := range X {
		for j := range X[i] {
			X[i][j] = rand.Float64()*(ub[0]-lb[0]) + lb[0]
		}
	}
	return X
}

// Function to divide the population: takes the original population 'X' and the number of population 'k'
func dividePopulation(X [][]float64, k int) [][][]float64 {
	totalSize := len(X)
	baseSubPopSize := totalSize / k
	remainder := totalSize % k

	Xsub := make([][][]float64, k)

	startIndex := 0
	for i := 0; i < k; i++ {
		subPopSize := baseSubPopSize
		if remainder > 0 { // In case the division is not even
			subPopSize++ // Add one of the remaining individuals to this sub-population
			remainder--
		}
		Xsub[i] = X[startIndex : startIndex+subPopSize]
		startIndex += subPopSize
	}

	return Xsub
}

func main() {

	kickStart := time.Now()
	rand.Seed(time.Now().UnixNano())

	N := 600   // Population
	K := 10    // Number of sub-population
	dim := 500 // Dimension
	T := 500

	lb := []float64{-100.0}
	ub := []float64{100.0}

	// Initialize population
	X := initializePopulation(N, dim, lb, ub)
	// Divide population
	Xpop := dividePopulation(X, K)

	BestFitness := math.Inf(1)
	BestPos := make([]float64, dim)
	GlobalPos := make([]float64, dim)
	globalCov := make([]float64, T)
	fitnessF := make([][]float64, len(Xpop))

	mutex := &sync.Mutex{}

	for i, subPop := range Xpop {
		wg.Add(1)
		fitnessF[i] = make([]float64, len(subPop)) // Initialize fitnessF for each sub-population

		go func(index int, subPop [][]float64) {
			defer wg.Done()

			localBestFitness := math.Inf(1)
			localBestPos := make([]float64, dim)

			for j, individual := range subPop {
				fitness := F6(individual)
				fitnessF[index][j] = fitness // Update the fitnessF slice

				if fitness < localBestFitness {
					localBestFitness = fitness
					copy(localBestPos, individual)
				}
			}

			mutex.Lock()
			if localBestFitness < BestFitness {
				BestFitness = localBestFitness
				copy(BestPos, localBestPos)
			}
			mutex.Unlock()

		}(i, subPop)
	}

	wg.Wait() // Wait for all goroutines to finish

	//Update Global
	copy(GlobalPos, BestPos)
	GlobalFitness := BestFitness

	Xf := make([]float64, dim) // For Xshade -- array for the cave
	Xfood := make([]float64, dim)

	// Initialize new positions for each sub-population
	Xnew := make([][][]float64, len(Xpop))
	for i := range Xpop {
		Xnew[i] = make([][]float64, len(Xpop[i]))
		for j := range Xpop[i] {
			Xnew[i][j] = make([]float64, dim)
		}
	}

	for t := 0; t < T; t++ {
		C := 2 - (float64(t) / float64(T))
		tmp := rand.Float64()*15 + 20

		for i := 0; i < dim; i++ {
			Xf[i] = (BestPos[i] + GlobalPos[i]) / 2
		}
		copy(Xfood, BestPos)

		for i, subPop := range Xpop {
			wg.Add(1)
			go func(i int, subPop [][]float64) {
				defer wg.Done()

				localTmp := tmp // Local copy of tmp for each goroutine

				for j := range subPop {
					if localTmp > 30 {
						if rand.Float64() < 0.5 {
							for k := 0; k < dim; k++ {
								Xnew[i][j][k] = subPop[j][k] + C*rand.Float64()*(Xf[k]-subPop[j][k])
							}
						} else { // Competition Stage
							for k := 0; k < dim; k++ {
								z := rand.Intn(len(subPop))
								Xnew[i][j][k] = subPop[j][k] - subPop[z][k] + Xf[k]
							}
						}
					} else { // Foraging Stage
						P := 3 * rand.Float64() * fitnessF[i][j] / F6(Xfood)
						if P > 2 {
							for k := 0; k < dim; k++ {
								Xfood[k] *= math.Exp(-1 / P)
								Xnew[i][j][k] = subPop[j][k] + math.Cos(2*math.Pi*rand.Float64())*Xfood[k]*p_obj(localTmp) - math.Sin(2*math.Pi*rand.Float64())*Xfood[k]*p_obj(localTmp)
							}
						} else {
							for k := 0; k < dim; k++ {
								Xnew[i][j][k] = (subPop[j][k]-Xfood[k])*p_obj(localTmp) + p_obj(localTmp)*rand.Float64()*subPop[j][k]
							}
						}
					}
				}

				// Boundary condition checks
				for j := range subPop {
					for k := 0; k < dim; k++ {
						if len(ub) == 1 {
							Xnew[i][j][k] = math.Min(ub[0], Xnew[i][j][k])
							Xnew[i][j][k] = math.Max(lb[0], Xnew[i][j][k])
						} else {
							Xnew[i][j][k] = math.Min(ub[k], Xnew[i][j][k])
							Xnew[i][j][k] = math.Max(lb[k], Xnew[i][j][k])
						}
					}
				}

				for j := range subPop {
					// Evaluate the fitness of the new position
					newFitness := F6(Xnew[i][j])

					// Update local best fitness
					mutex.Lock()
					if newFitness < BestFitness {
						BestFitness = newFitness
						copy(BestPos, Xnew[i][j])
					}
					mutex.Unlock()

					// Update individual position if new fitness is better
					if newFitness < fitnessF[i][j] {
						fitnessF[i][j] = newFitness
						copy(subPop[j], Xnew[i][j])
					}
				}

			}(i, subPop)
		}

		wg.Wait() // Wait for all goroutines to finish

		// Global Update
		copy(GlobalPos, Xpop[0][0])
		GlobalFitness = F6(GlobalPos)

		// Iterate over all sub-populations and individuals to find the global best
		for _, subPop := range Xpop {
			for _, individual := range subPop {
				individualFitness := F6(individual)
				if individualFitness < GlobalFitness {
					GlobalFitness = individualFitness
					copy(GlobalPos, individual)
				}
			}
		}
		// Store the global best fitness at this iteration
		globalCov[t] = GlobalFitness
	}

	fmt.Println("Best Fitness: ", BestFitness)
	fmt.Println("Executed in: ", time.Since(kickStart))

}
