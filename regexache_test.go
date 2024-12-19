package regexache

import (
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
	pattern := `(\d{3}-\d{2}-\d{4})|($begin:math:text$\\d{3}$end:math:text$\s\d{3}-\d{4})` // US phone numbers (basic format)
	regex1 := MustCompile(pattern)
	regex2 := MustCompile(pattern)

	if regex1 != regex2 {
		t.Errorf("Expected cached regex instances to be the same for pattern: %s", pattern)
	}
}

func BenchmarkMustCompile(b *testing.B) {
	// Define patterns for consistency
	literalPattern := "abc"
	alphaNumericPattern := `^[a-zA-Z0-9_]+$`
	ipPattern := `^(\d{1,3}\.){3}\d{1,3}$`
	datePattern := `^\d{4}-\d{2}-\d{2}$`
	colorCodePattern := `^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`
	phoneNumberPattern := `(\d{3}-\d{2}-\d{4})|($begin:math:text$\\d{3}$end:math:text$\s\d{3}-\d{4})`

	// Benchmark for literal pattern
	b.Run("CachingDisabled_LiteralPattern", func(b *testing.B) {
		caching = false
		for i := 0; i < b.N; i++ {
			MustCompile(literalPattern)
		}
	})

	b.Run("CachingEnabled_LiteralPattern", func(b *testing.B) {
		caching = true
		MustCompile(literalPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(literalPattern)
		}
	})

	// Benchmark for alphanumeric pattern
	b.Run("CachingDisabled_AlphaNumericPattern", func(b *testing.B) {
		caching = false
		for i := 0; i < b.N; i++ {
			MustCompile(alphaNumericPattern)
		}
	})

	b.Run("CachingEnabled_AlphaNumericPattern", func(b *testing.B) {
		caching = true
		MustCompile(alphaNumericPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(alphaNumericPattern)
		}
	})

	// Benchmark for IP pattern
	b.Run("CachingDisabled_IPPattern", func(b *testing.B) {
		caching = false
		for i := 0; i < b.N; i++ {
			MustCompile(ipPattern)
		}
	})

	b.Run("CachingEnabled_IPPattern", func(b *testing.B) {
		caching = true
		MustCompile(ipPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(ipPattern)
		}
	})

	// Benchmark for date pattern
	b.Run("CachingDisabled_DatePattern", func(b *testing.B) {
		caching = false
		for i := 0; i < b.N; i++ {
			MustCompile(datePattern)
		}
	})

	b.Run("CachingEnabled_DatePattern", func(b *testing.B) {
		caching = true
		MustCompile(datePattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(datePattern)
		}
	})

	// Benchmark for color code pattern
	b.Run("CachingDisabled_ColorCodePattern", func(b *testing.B) {
		caching = false
		for i := 0; i < b.N; i++ {
			MustCompile(colorCodePattern)
		}
	})

	b.Run("CachingEnabled_ColorCodePattern", func(b *testing.B) {
		caching = true
		MustCompile(colorCodePattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(colorCodePattern)
		}
	})

	// Benchmark for phone number pattern
	b.Run("CachingDisabled_PhoneNumberPattern", func(b *testing.B) {
		caching = false
		for i := 0; i < b.N; i++ {
			cache = sync.Map{} // reset global cache
			MustCompile(phoneNumberPattern)
		}
	})

	b.Run("CachingEnabled_PhoneNumberPattern", func(b *testing.B) {
		caching = true
		MustCompile(phoneNumberPattern) // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile(phoneNumberPattern)
		}
	})
}
