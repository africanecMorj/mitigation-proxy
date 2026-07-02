package runtime

import (
	"testing"
	"strconv"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
)

func TestConcurrentReload(t *testing.T) {
	cfg := &config.Config{}

	rt := New()

	rt.Build(cfg)

    done := make(chan struct{})

    go func() {
        for i := 0; i < 1000; i++ {
            rt.Reload(cfg)
        }
        close(done)
    }()

    for i := 0; i < 100000; i++ {
        clusters := rt.clusters.Load()

        for _, lb := range clusters.clusters {
            _ = lb.Backends()
        }
    }

    <-done
}

func BenchmarkRegister(b *testing.B) {
	rt := New()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rt.Register(
			string(rune(i)),
			nil,
		)
	}
}

func BenchmarkGet(b *testing.B) {
	rt := New()

	for i := 0; i < 1000; i++ {
		rt.loops.Load().loops[string(rune(i))] = nil
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rt.Get(string(rune(i % 1000)))
	}
}

func BenchmarkRegisterClusters(b *testing.B) {
	rt := New()

	clusters := make(map[string]balancers.Balancer, 1000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rt.RegisterClusters(clusters)
	}
}

func BenchmarkCloneMap100(b *testing.B) {
	src := make(map[int]int, 100)

	for i := 0; i < 100; i++ {
		src[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cloneMap(src)
	}
}

func BenchmarkCloneMap1000(b *testing.B) {
	src := make(map[int]int, 1000)

	for i := 0; i < 1000; i++ {
		src[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cloneMap(src)
	}
}

func BenchmarkCloneMap10000(b *testing.B) {
	src := make(map[int]int, 10000)

	for i := 0; i < 10000; i++ {
		src[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cloneMap(src)
	}
}

func BenchmarkGetParallel(b *testing.B) {
	rt := New()

	loops := make(map[string]*transport.Transport)

	for i := 0; i < 1000; i++ {
		loops[strconv.Itoa(i)] = nil
	}

	rt.loops.Store(&loopSnapshot{
		loops: loops,
	})

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			rt.Get(strconv.Itoa(i % 1000))
			i++
		}
	})
}

func BenchmarkRegisterWhileReading(b *testing.B) {
	rt := New()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				rt.Get("test")
			}
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rt.Register(strconv.Itoa(i), nil)
	}

	close(done)
}