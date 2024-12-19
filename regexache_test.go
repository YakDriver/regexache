package regexache

import (
	"sync"
	"testing"
)

func TestMustCompile_basic(t *testing.T) {
	// Clear cache before testing
	cache = sync.Map{}

	regex1 := MustCompile("abc")
	regex2 := MustCompile("abc")

	if regex1 != regex2 {
		t.Errorf("Expected cached regex instances to be the same, but got different instances")
	}
}

func TestMustCompile_cachingDisabled(t *testing.T) {
	// Set environment variable to disable caching
	t.Setenv(REGEXACHE_OFF, "1")
	// Clear cache before testing
	cache = sync.Map{}

	regex1 := MustCompile("abc")
	regex2 := MustCompile("abc")

	// Check that different pointers are returned,
	// if something went wrong MustCompile would've panicked already.
	if regex1 == regex2 {
		t.Errorf("Expected different regex objects when caching is disabled, but got the same instance")
	}
}

func BenchmarkMustCompile(b *testing.B) {
	// Benchmark with an empty cache each time (no preload)
	b.Run("Caching disabled", func(b *testing.B) {
		b.Setenv(REGEXACHE_OFF, "1")

		for i := 0; i < b.N; i++ {
			cache = sync.Map{} // reset global cache
			MustCompile("abc")
		}
	})

	b.Run("Preload disabled", func(b *testing.B) {
		b.Setenv(REGEXACHE_PRELOAD_OFF, "1")

		for i := 0; i < b.N; i++ {
			cache = sync.Map{} // reset global cache
			MustCompile("abc")
		}
	})

	// Benchmark with a populated cache (preloaded)
	b.Run("Caching enabled", func(b *testing.B) {
		cache = sync.Map{}
		MustCompile("abc") // preload the cache

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			MustCompile("abc")
		}
	})
}
