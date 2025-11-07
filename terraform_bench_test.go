package regexache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// Simulate AWS provider regex patterns - mix of common and unique patterns
var awsPatterns = []string{
	// Common validation patterns (high reuse)
	`^[a-zA-Z0-9._-]+$`,
	`^[a-zA-Z][a-zA-Z0-9._-]*$`,
	`^arn:aws:.*`,
	`^[0-9]+$`,
	`^[a-z0-9-]+$`,
	
	// Resource-specific patterns (medium reuse)
	`^sg-[0-9a-f]{8,17}$`,
	`^vpc-[0-9a-f]{8,17}$`,
	`^subnet-[0-9a-f]{8,17}$`,
	`^i-[0-9a-f]{8,17}$`,
	`^vol-[0-9a-f]{8,17}$`,
	
	// Complex validation patterns (low reuse)
	`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}(?:/[0-9]{1,2})?$`,
	`^[a-zA-Z0-9+/]*={0,2}$`,
	`^[A-Z0-9]{20}$`,
	`^[a-zA-Z0-9+/]{40}$`,
	`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$`,
}

// Generate unique patterns to simulate large provider diversity
func generateUniquePatterns(count int) []string {
	patterns := make([]string, count)
	for i := 0; i < count; i++ {
		patterns[i] = fmt.Sprintf(`^resource-%d-[a-zA-Z0-9]{%d}$`, i, 8+i%10)
	}
	return patterns
}

// BenchmarkTerraformUsagePattern simulates how Terraform actually uses regexache
func BenchmarkTerraformUsagePattern(b *testing.B) {
	// Generate patterns similar to terraform-provider-aws scale
	commonPatterns := awsPatterns
	uniquePatterns := generateUniquePatterns(4500) // Simulate AWS provider scale
	
	b.Run("SmallConfig_CachingEnabled", func(b *testing.B) {
		SetCaching(true)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate small Terraform config - mostly common patterns
			for j := 0; j < 50; j++ {
				pattern := commonPatterns[j%len(commonPatterns)]
				MustCompile(pattern)
			}
		}
	})
	
	b.Run("SmallConfig_CachingDisabled", func(b *testing.B) {
		SetCaching(false)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for j := 0; j < 50; j++ {
				pattern := commonPatterns[j%len(commonPatterns)]
				MustCompile(pattern)
			}
		}
	})
	
	b.Run("LargeConfig_CachingEnabled", func(b *testing.B) {
		SetCaching(true)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate large Terraform config - mix of common and unique
			for j := 0; j < 1000; j++ {
				var pattern string
				if j%10 < 3 { // 30% common patterns
					pattern = commonPatterns[j%len(commonPatterns)]
				} else { // 70% unique patterns
					pattern = uniquePatterns[j%len(uniquePatterns)]
				}
				MustCompile(pattern)
			}
		}
	})
	
	b.Run("LargeConfig_CachingDisabled", func(b *testing.B) {
		SetCaching(false)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				var pattern string
				if j%10 < 3 {
					pattern = commonPatterns[j%len(commonPatterns)]
				} else {
					pattern = uniquePatterns[j%len(uniquePatterns)]
				}
				MustCompile(pattern)
			}
		}
	})
	
	b.Run("WorstCase_AllUnique_CachingEnabled", func(b *testing.B) {
		SetCaching(true)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Worst case: all patterns are unique (no cache benefit)
			for j := 0; j < 500; j++ {
				pattern := uniquePatterns[j%len(uniquePatterns)]
				MustCompile(pattern)
			}
		}
	})
	
	b.Run("WorstCase_AllUnique_CachingDisabled", func(b *testing.B) {
		SetCaching(false)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for j := 0; j < 500; j++ {
				pattern := uniquePatterns[j%len(uniquePatterns)]
				MustCompile(pattern)
			}
		}
	})
}

// BenchmarkMemoryPressure tests memory usage under different scenarios
func BenchmarkMemoryPressure(b *testing.B) {
	patterns := generateUniquePatterns(10000)
	
	b.Run("CacheGrowth_Enabled", func(b *testing.B) {
		SetCaching(true)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate cache growing with unique patterns
			for j := 0; j < 1000; j++ {
				MustCompile(patterns[j])
			}
		}
	})
	
	b.Run("NoCache_Disabled", func(b *testing.B) {
		SetCaching(false)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				MustCompile(patterns[j])
			}
		}
	})
}

// BenchmarkPatternReuse tests different reuse scenarios
func BenchmarkPatternReuse(b *testing.B) {
	basePattern := `^[a-zA-Z0-9._-]+$`
	
	scenarios := []struct {
		name      string
		reuseRate float64 // Probability of reusing base pattern vs unique
	}{
		{"HighReuse_90pct", 0.9},
		{"MediumReuse_50pct", 0.5},
		{"LowReuse_10pct", 0.1},
		{"NoReuse_0pct", 0.0},
	}
	
	for _, scenario := range scenarios {
		b.Run(scenario.name+"_CachingEnabled", func(b *testing.B) {
			SetCaching(true)
			rand.Seed(time.Now().UnixNano())
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				for j := 0; j < 1000; j++ {
					if rand.Float64() < scenario.reuseRate {
						MustCompile(basePattern)
					} else {
						MustCompile(fmt.Sprintf(`^unique-%d-[a-z]+$`, j))
					}
				}
			}
		})
		
		b.Run(scenario.name+"_CachingDisabled", func(b *testing.B) {
			SetCaching(false)
			rand.Seed(time.Now().UnixNano())
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				for j := 0; j < 1000; j++ {
					if rand.Float64() < scenario.reuseRate {
						MustCompile(basePattern)
					} else {
						MustCompile(fmt.Sprintf(`^unique-%d-[a-z]+$`, j))
					}
				}
			}
		})
	}
}

// BenchmarkColdStart simulates Terraform's cold start behavior
func BenchmarkColdStart(b *testing.B) {
	patterns := append(awsPatterns, generateUniquePatterns(100)...)
	
	b.Run("ColdStart_WithPreload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Reset cache to simulate cold start
			cache = sync.Map{}
			SetCaching(true)
			
			b.StartTimer()
			// Simulate initial pattern compilation burst
			for _, pattern := range patterns {
				MustCompile(pattern)
			}
			b.StopTimer()
		}
	})
	
	b.Run("ColdStart_NoCache", func(b *testing.B) {
		SetCaching(false)
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for _, pattern := range patterns {
				MustCompile(pattern)
			}
		}
	})
}
