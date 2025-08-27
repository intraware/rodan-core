package tinylfu_test

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/go-tinylfu"
)

func TestCache(t *testing.T) {
	cache := tinylfu.New(1e3, 10e3)
	keys := []string{"one", "two", "three"}

	for _, key := range keys {
		cache.Set(&tinylfu.Item{
			Key:   key,
			Value: key,
		})

		got, ok := cache.Get(key)
		require.True(t, ok)
		require.Equal(t, key, got)
	}

	for _, key := range keys {
		got, ok := cache.Get(key)
		require.True(t, ok)
		require.Equal(t, key, got)

		cache.Set(&tinylfu.Item{
			Key:   key,
			Value: key + key,
		})
	}

	for _, key := range keys {
		got, ok := cache.Get(key)
		require.True(t, ok)
		require.Equal(t, key+key, got)
	}

	for _, key := range keys {
		cache.Del(key)
	}

	for _, key := range keys {
		_, ok := cache.Get(key)
		require.False(t, ok)
	}
}

func TestOOM(t *testing.T) {
	keys := make([]string, 10000)
	for i := range keys {
		keys[i] = randWord()
	}

	cache := tinylfu.New(1e3, 10e3)

	for i := 0; i < 5e6; i++ {
		key := keys[i%len(keys)]
		cache.Set(&tinylfu.Item{
			Key:   key,
			Value: key,
		})
	}
}

func TestCorruptionOnExpiry(t *testing.T) {
	const size = 50000

	strFor := func(i int) string {
		return fmt.Sprintf("a string %d", i)
	}
	keyName := func(i int) string {
		return fmt.Sprintf("key-%00000d", i)
	}

	mycache := tinylfu.New(1000, 10000)
	// Put a bunch of stuff in the cache with a TTL of 1 second
	for i := 0; i < size; i++ {
		key := keyName(i)
		mycache.Set(&tinylfu.Item{
			Key:      key,
			Value:    []byte(strFor(i)),
			ExpireAt: time.Now().Add(time.Second),
		})
	}

	// Read stuff for a bit longer than the TTL - that's when the corruption occurs
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := ctx.Done()
loop:
	for {
		select {
		case <-done:
			// this is expected
			break loop
		default:
			i := rand.Intn(size)
			key := keyName(i)

			b, ok := mycache.Get(key)
			if !ok {
				continue loop
			}

			got := string(b.([]byte))
			expected := strFor(i)
			if got != expected {
				t.Fatalf("expected=%q got=%q key=%q", expected, got, key)
			}
		}
	}
}

func randWord() string {
	buf := make([]byte, 64)
	io.ReadFull(cryptorand.Reader, buf)
	return string(buf)
}

func TestAddAlreadyInCache(t *testing.T) {
	c := tinylfu.New(100, 10000)

	c.Set(&tinylfu.Item{
		Key:   "foo",
		Value: "bar",
	})

	val, _ := c.Get("foo")
	if val.(string) != "bar" {
		t.Errorf("c.Get(foo)=%q, want %q", val, "bar")
	}

	c.Set(&tinylfu.Item{
		Key:   "foo",
		Value: "baz",
	})

	val, _ = c.Get("foo")
	if val.(string) != "baz" {
		t.Errorf("c.Get(foo)=%q, want %q", val, "baz")
	}
}
