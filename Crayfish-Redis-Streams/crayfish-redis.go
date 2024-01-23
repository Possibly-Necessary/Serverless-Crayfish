// This code runs perfectly fine (tested)

package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/go-redis/redis"
)

// Benchmark function F6 - Boundary range [-100,100] --> will remove this and implement benchmark function
func F6(x []float64) float64 {
	var o float64
	for _, value := range x {
		o += math.Pow(math.Abs(value+0.5), 2)
	}
	return o
}

// Funtion to initialize and divide the population
func initializePopulation(N, dim, k int, lb, ub []float64) [][][]float64 { // Instead of returning ([]byte, error)

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

	// Split the population based on k
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

// Equation 4: Mathimatical model of crayfish intake
func p_obj(x float64) float64 {
	return 0.2 * (1 / (math.Sqrt(2*math.Pi) * 3)) * math.Exp(-math.Pow(x-25, 2)/(2*math.Pow(3, 2)))
}

func crayfish(T int, lb, ub []float64, X [][]float64) (x, y []float64) { // return bestFit, bestPos

	N := len(X) // size of the sub-population
	dim := len(X[0])

	var (
		globalCov   []float64 = make([]float64, T) // zero row vector of size T
		BestFitness           = math.Inf(1)
		BestPos     []float64 = make([]float64, dim)
		fitnessF    []float64 = make([]float64, N)
		GlobalPos   []float64 = make([]float64, dim)
	)

	for i := 0; i < N; i++ {
		fitnessF[i] = F6(X[i]) // Get the fitness value from the benchmark function
		if fitnessF[i] < BestFitness {
			BestFitness = fitnessF[i]
			copy(BestPos, X[i])
		}
	}

	// Update best position to Global position
	copy(GlobalPos, BestPos)
	GlobalFitness := BestFitness

	Xf := make([]float64, dim) // For Xshade -- array for the cave
	Xfood := make([]float64, dim)

	Xnew := make([][]float64, N) // Initializing a 2d array
	for i := 0; i < N; i++ {
		Xnew[i] = make([]float64, dim)
	}

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
				P := 3 * rand.Float64() * fitnessF[i] / F6(Xfood)
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
		GlobalFitness = F6(GlobalPos)

		for i := 0; i < N; i++ {
			NewFitness := F6(Xnew[i])
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

		t++
	}

	return BestPos, globalCov
}

// Update the publish function to accept optimization results
func publishOptimizationResults(client *redis.Client, bestPos []float64, bestFit float64, globalCov []float64) error {
	log.Println("Publishing optimization results to Redis")

	// Convert bestPos and globalCov to strings for Redis
	bestPosStr := fmt.Sprintf("%v", bestPos)
	globalCovStr := fmt.Sprintf("%v", globalCov)

	err := client.XAdd(&redis.XAddArgs{ // XAdd method to add data to the Redis stream
		Stream: "optimization_results", // Stream name
		ID:     "",
		Values: map[string]interface{}{
			"bestPosition": bestPosStr,
			"bestFitness":  bestFit,
			"globalCov":    globalCovStr,
		},
	}).Err()

	return err
}

func main() {

	// Crayfish Initialization
	N, K, T := 20, 4, 4
	lb := []float64{-100.0}
	ub := []float64{100.0}
	dim := 3


	// Redis Connection stuff
	log.Println("Publisher Started")

	redisClient := redis.NewClient(&redis.Options{ // Initialize Redit client and connect to the server at `127.0.0.1:6379'
		Addr: fmt.Sprintf("%s:%s", "127.0.0.1", "6379"),
	})
	_, err := redisClient.Ping().Result() // Ping Redis server to check the connection
	if err != nil {
		log.Fatal("Unbale to connect to Redis", err)
	}

	log.Println("Connected to Redis server")

	// intialize the split population of crayfish
	X := initializePopulation(N, dim, K, lb, ub)

	// Process each sub-population
	for _, subPop := range X { // For each sub-population in X
		bestPos, globalCov := crayfish(T, lb, ub, subPop) // This part will be parallel in Nuclio
		bestFit := F6(bestPos)
		// Publish result to Redis
		err := publishOptimizationResults(redisClient, bestPos, bestFit, globalCov)
		if err != nil {
			log.Fatal("Failed to publish results to Redis.\n", err)
		}

	}

}
