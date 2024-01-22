package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/rs/xid"
)

func F6(x []float64) float64 {
	var o float64
	for _, value := range x {
		o += math.Pow(math.Abs(value+0.5), 2)
	}
	return o
}

func parseStringSliceToFloatSlice(s string) ([]float64, error) {
	// Remove the brackets [] from the string
	s = strings.Trim(s, "[]")

	// Split the string into a slice of strings
	strs := strings.Split(s, " ")

	// Convert each string in the slice to a float64
	var result []float64
	for _, str := range strs {
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

func parseOptimizationResults(values map[string]interface{}) (bestFit float64, bestPos []float64, globalCov []float64) {
	// Parse the values map to extract bestFit, bestPos, and globalCov
	// Parse bestFit
	bestFitStr := values["bestFitness"].(string)
	bestFit, _ = strconv.ParseFloat(bestFitStr, 64)

	// Parse bestPos
	bestPosStr := values["bestPosition"].(string)
	bestPos, _ = parseStringSliceToFloatSlice(bestPosStr)

	// Parse GlobalCov
	// Parse globalCov
	globalCovStr := values["globalCov"].(string)
	globalCov, _ = parseStringSliceToFloatSlice(globalCovStr)
	return
}

func updateOverallResults(overallBestFit *float64, overallBestPos *[]float64, overallGlobalCov *[]float64, bestFit float64, bestPos []float64, globalCov []float64) {
	// Update overall best fitness and position
	if bestFit < *overallBestFit {
		*overallBestFit = bestFit
		*overallBestPos = make([]float64, len(bestPos))
		copy(*overallBestPos, bestPos)
	}

	// Accumulate global convergence values
	if *overallGlobalCov == nil {
		*overallGlobalCov = make([]float64, len(globalCov))
	}
	for i, cov := range globalCov {
		(*overallGlobalCov)[i] += cov
	}
}

func main() {

	log.Println("Consumer started")

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", "127.0.0.1", "6379"),
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal("Unbale to connect to Redis", err)
	}

	log.Println("Connected to Redis server")

	subject := "optimization_results"
	consumersGroup := "optimization-consumer-group"

	err = redisClient.XGroupCreate(subject, consumersGroup, "0").Err()
	if err != nil {
		log.Println(err)
	}

	uniqueID := xid.New().String()

	var (
		overallBestFit   = math.Inf(1)
		overallBestPos   []float64
		overallGlobalCov []float64
	)
	messageCount := 0
	totalWorkers := 4

	for messageCount < totalWorkers {
		entries, err := redisClient.XReadGroup(&redis.XReadGroupArgs{
			Group:    consumersGroup,
			Consumer: uniqueID, // Use uniqueID here
			Streams:  []string{subject, ">"},
			Count:    2,
			Block:    0,
			NoAck:    false,
		}).Result()
		if err != nil {
			log.Fatal(err)
		}

		for _, message := range entries[0].Messages {
			bestFit, bestPos, globalCov := parseOptimizationResults(message.Values)
			updateOverallResults(&overallBestFit, &overallBestPos, &overallGlobalCov, bestFit, bestPos, globalCov)
			redisClient.XAck(subject, consumersGroup, message.ID)
			messageCount++
		}
	}

	// Average the global convergence values
	for i := range overallGlobalCov {
		overallGlobalCov[i] /= float64(totalWorkers)
	}

	fmt.Println("Overall Best Fitness:", overallBestFit)
	fmt.Println("Overall Best Position:", overallBestPos)
	fmt.Println("Overall Global Convergence:", overallGlobalCov)

}
