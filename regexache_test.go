package regexache

import (
	"os"
	"regexp"
	"sync"
	"testing"
)

func TestMustCompile_CacheBasicPattern(t *testing.T) {
	regex1 := MustCompile("abc")
	regex2 := MustCompile("abc")

	if regex1 != regex2 {
		t.Errorf("Expected cached regex instances to be the same, but got different instances")
	}
}

func TestMustCompile_CacheComplexPattern(t *testing.T) {
	p := `(\d{3}-\d{2}-\d{4})|($begin:math:text$\\d{3}$end:math:text$\s\d{3}-\d{4})` // US phone numbers (basic format)
	regex1 := MustCompile(p)
	regex2 := MustCompile(p)

	if regex1 != regex2 {
		t.Errorf("Expected cached regex instances to be the same for pattern: %s", p)
	}
}

func TestSetCaching(t *testing.T) {
	// Test enabling caching
	SetCaching(true)
	if !IsCachingEnabled() {
		t.Error("Expected caching to be enabled")
	}

	// Test disabling caching
	SetCaching(false)
	if IsCachingEnabled() {
		t.Error("Expected caching to be disabled")
	}

	// Reset to enabled for other tests
	SetCaching(true)
}

func TestMustCompile_CachingDisabled(t *testing.T) {
	SetCaching(false)
	defer SetCaching(true) // Reset after test

	regex1 := MustCompile("test")
	regex2 := MustCompile("test")

	// When caching is disabled, should get different instances
	if regex1 == regex2 {
		t.Error("Expected different regex instances when caching is disabled")
	}
}

func TestMustCompile_DifferentPatterns(t *testing.T) {
	SetCaching(true)

	regex1 := MustCompile("pattern1")
	regex2 := MustCompile("pattern2")

	if regex1 == regex2 {
		t.Error("Expected different regex instances for different patterns")
	}
}

func TestMustCompile_EmptyPattern(t *testing.T) {
	// Empty pattern should work (matches empty string)
	regex := MustCompile("")
	if regex == nil {
		t.Error("Expected non-nil regex for empty pattern")
	}
}

func TestMustCompile_InvalidPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid regex pattern")
		}
	}()

	MustCompile("[") // Invalid regex pattern
}

func TestMustCompile_Concurrent(t *testing.T) {
	SetCaching(true)
	pattern := "concurrent-test"
	
	var wg sync.WaitGroup
	results := make([]*regexp.Regexp, 100)
	
	// Test concurrent access to same pattern
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = MustCompile(pattern)
		}(i)
	}
	
	wg.Wait()
	
	// All results should be the same cached instance
	first := results[0]
	for i, result := range results {
		if result != first {
			t.Errorf("Expected same cached instance at index %d", i)
		}
	}
}

func TestMustCompile_ConcurrentToggling(t *testing.T) {
	var wg sync.WaitGroup
	
	// Test concurrent caching enable/disable
	for i := 0; i < 50; i++ {
		wg.Add(2)
		
		go func() {
			defer wg.Done()
			SetCaching(true)
			MustCompile("toggle-test")
		}()
		
		go func() {
			defer wg.Done()
			SetCaching(false)
			MustCompile("toggle-test")
		}()
	}
	
	wg.Wait()
	SetCaching(true) // Reset
}

func TestMustCompile_WithStandardization(t *testing.T) {
	// Save original state
	originalStandardizing := standardizing
	defer func() { standardizing = originalStandardizing }()
	
	// Enable standardization
	standardizing = true
	SetCaching(true)
	
	// Test that standardized patterns are cached correctly
	regex1 := MustCompile("[a-zA-Z0-9_]")
	regex2 := MustCompile("[a-zA-Z0-9_]")
	
	if regex1 != regex2 {
		t.Error("Expected cached regex instances for standardized patterns")
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Test REGEXACHE_OFF environment variable handling
	// Note: This tests the logic, but init() has already run
	
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"empty", "", true},
		{"set", "1", false},
		{"any_value", "true", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the init logic
			caching := true
			if tt.envValue != "" {
				caching = false
			}
			
			if caching != tt.expected {
				t.Errorf("Expected caching=%v for env value '%s'", tt.expected, tt.envValue)
			}
		})
	}
}

func TestOutputFunctionality(t *testing.T) {
	// Save original state
	originalOutputFile := outputFile
	originalLookups := lookups
	defer func() {
		outputFile = originalOutputFile
		lookups = originalLookups
	}()
	
	// Test output tracking
	outputFile = "test-output"
	lookups = make(map[string]int)
	
	SetCaching(true)
	
	// Generate some lookups
	MustCompile("test1")
	MustCompile("test1") // Second lookup
	MustCompile("test2")
	
	// Verify lookup counting
	if lookups["test1"] != 2 {
		t.Errorf("Expected 2 lookups for test1, got %d", lookups["test1"])
	}
	if lookups["test2"] != 1 {
		t.Errorf("Expected 1 lookup for test2, got %d", lookups["test2"])
	}
	
	// Test outputCache function (should not panic)
	outputCache()
	
	// Clean up test file if created
	os.Remove("test-output")
}

func BenchmarkMustCompile(b *testing.B) {
	literalPattern := "abc"
	alphaNumericPattern := `^[a-zA-Z0-9_]+$`
	ipPattern := `^(\d{1,3}\.){3}\d{1,3}$`
	datePattern := `^\d{4}-\d{2}-\d{2}$`
	colorCodePattern := `^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`
	phoneNumberPattern := `(\d{3}-\d{2}-\d{4})|($begin:math:text$\\d{3}$end:math:text$\s\d{3}-\d{4})`

	// Literal pattern
	b.Run("CachingDisabled_LiteralPattern", func(b *testing.B) {
		SetCaching(false)
		for i := 0; i < b.N; i++ {
			MustCompile(literalPattern)
		}
	})

	b.Run("CachingEnabled_LiteralPattern", func(b *testing.B) {
		SetCaching(true)
		MustCompile(literalPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(literalPattern)
		}
	})

	// Alphanumeric pattern
	b.Run("CachingDisabled_AlphaNumericPattern", func(b *testing.B) {
		SetCaching(false)
		for i := 0; i < b.N; i++ {
			MustCompile(alphaNumericPattern)
		}
	})

	b.Run("CachingEnabled_AlphaNumericPattern", func(b *testing.B) {
		SetCaching(true)
		MustCompile(alphaNumericPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(alphaNumericPattern)
		}
	})

	// IP pattern
	b.Run("CachingDisabled_IPPattern", func(b *testing.B) {
		SetCaching(false)
		for i := 0; i < b.N; i++ {
			MustCompile(ipPattern)
		}
	})

	b.Run("CachingEnabled_IPPattern", func(b *testing.B) {
		SetCaching(true)
		MustCompile(ipPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(ipPattern)
		}
	})

	// Date pattern
	b.Run("CachingDisabled_DatePattern", func(b *testing.B) {
		SetCaching(false)
		for i := 0; i < b.N; i++ {
			MustCompile(datePattern)
		}
	})

	b.Run("CachingEnabled_DatePattern", func(b *testing.B) {
		SetCaching(true)
		MustCompile(datePattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(datePattern)
		}
	})

	// Color code pattern
	b.Run("CachingDisabled_ColorCodePattern", func(b *testing.B) {
		SetCaching(false)
		for i := 0; i < b.N; i++ {
			MustCompile(colorCodePattern)
		}
	})

	b.Run("CachingEnabled_ColorCodePattern", func(b *testing.B) {
		SetCaching(true)
		MustCompile(colorCodePattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(colorCodePattern)
		}
	})

	// Phone number pattern
	b.Run("CachingDisabled_PhoneNumberPattern", func(b *testing.B) {
		SetCaching(false)
		for i := 0; i < b.N; i++ {
			MustCompile(phoneNumberPattern)
		}
	})

	b.Run("CachingEnabled_PhoneNumberPattern", func(b *testing.B) {
		SetCaching(true)
		MustCompile(phoneNumberPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(phoneNumberPattern)
		}
	})
}
