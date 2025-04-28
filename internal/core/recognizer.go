package core

import (
	"math"
	"sort"
	"sync"
	"go-learning/internal/model"
	"go-learning/internal/storage"
)

type MatchResult struct {
	Person  model.Person
	Score float64
}

func cosine(a, b [256]float64) float64 {
	var dot, normA, normB float64
	for i := 0; i < 256; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0 // avoid division by zero
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func GetSimilarPersons(input [256]float64, storage *storage.MemoryStore, topN int) []MatchResult {
	all := storage.GetPersons() // Get all persons from the in-memory slice
	results := make(chan MatchResult, len(*all)) // Buffered so goroutines don't block

	var wg sync.WaitGroup
	for _, person := range *all {
		wg.Add(1)

		go func(p model.Person) {
			defer wg.Done()
			score := cosine(input, p.Features)
			results <- MatchResult{Person: p, Score: score}
		}(person)
	}

	wg.Wait()
	close(results)

	// Collect all match results
	var matches []MatchResult
	for res := range results {
		matches = append(matches, res)
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > topN {
		return matches[:topN]
	}
	return matches
}
