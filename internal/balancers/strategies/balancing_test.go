package strategies

import (
	"testing"
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

func TestLatency(t *testing.T) {
	expected, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)
	spoiled, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)

	expected.ObserveLatency(100 * time.Millisecond)
	spoiled.ObserveLatency(500 * time.Millisecond)

	bl := NewLeastConnections([]*health.Backend{
		expected,
		spoiled,
	})

	got := bl.Next()
	if got != expected {
		gotAddr := "<nil>"
		if got != nil {
			gotAddr = got.Address
		}

		t.Errorf("expected backend %p (%s), got %p (%s)",
			expected, expected.Address,
			got, gotAddr,
		)
		t.Fail()
	} 
}

func TestTTFB(t *testing.T) {
	expected, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)
	spoiled, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)

	expected.SetTTFB(100 * time.Millisecond)
	spoiled.SetTTFB(500 * time.Millisecond)

	bl := NewLeastConnections([]*health.Backend{
		expected,
		spoiled,
	})

	got := bl.Next()
	if got != expected {
		gotAddr := "<nil>"
		if got != nil {
			gotAddr = got.Address
		}

		t.Errorf("expected backend %p (%s), got %p (%s)",
			expected, expected.Address,
			got, gotAddr,
		)
		t.Fail()
	} 

}

func TestWeight(t *testing.T) {
	expected, _ := health.NewBackend("127.0.0.1:3000",0,100,nil,nil)
	spoiled, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)

	expected.SetState(health.Healthy)
	spoiled.SetState(health.Healthy)

	bl := NewLeastConnections([]*health.Backend{
		expected,
		spoiled,
	})

	got := bl.Next()
	if got != expected {
		gotAddr := "<nil>"
		if got != nil {
			gotAddr = got.Address
		}

		t.Errorf("expected backend %p (%s), got %p (%s)",
			expected, expected.Address,
			got, gotAddr,
		)
		t.Fail()
	}
}

func TestConnections(t *testing.T) {
	expected, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)
	spoiled, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)

	expected.ActiveConnections.Add(10)
	spoiled.ActiveConnections.Add(100)

	bl := NewLeastConnections([]*health.Backend{
		expected,
		spoiled,
	})

	got := bl.Next()
	if got != expected {
		gotAddr := "<nil>"
		if got != nil {
			gotAddr = got.Address
		}

		t.Errorf("expected backend %p (%s), got %p (%s)",
			expected, expected.Address,
			got, gotAddr,
		)
		t.Fail()
	}
}

func TestRoundRobin(t *testing.T) {
	first, _ := health.NewBackend("127.0.0.1:3000", 0, 1, nil, nil)
	second, _ := health.NewBackend("127.0.0.1:3001", 0, 1, nil, nil)

	bl := NewRoundRobin([]*health.Backend{
		first,
		second,
	})

	expected := []*health.Backend{
		first,
		second,
		first,
		second,
		first,
		second,
		first,
		second,
		first,
		second,
	}

	for i, want := range expected {
		got := bl.Next()

		if got != want {
			gotAddr := "<nil>"
			if got != nil {
				gotAddr = got.Address
			}

			t.Errorf(
				"iteration %d: expected %s, got %s",
				i,
				want.Address,
				gotAddr,
			)
			t.Fail()
		}
	}
}

func TestGeneralP2C(t *testing.T) {
	expected, _ := health.NewBackend("127.0.0.1:3000",0,10,nil,nil)
	spoiled, _ := health.NewBackend("127.0.0.1:3000",0,0,nil,nil)

	expected.ActiveConnections.Add(10)
	spoiled.ActiveConnections.Add(100)

	expected.ObserveLatency(100 * time.Millisecond)
	spoiled.ObserveLatency(500 * time.Millisecond)

	expected.SetTTFB(100 * time.Millisecond)
	spoiled.SetTTFB(500 * time.Millisecond)

	bl := NewP2C([]*health.Backend{
		expected,
		spoiled,
	})

	got := bl.Next()
	if got != expected {
		gotAddr := "<nil>"
		if got != nil {
			gotAddr = got.Address
		}

		t.Errorf("expected backend %p (%s), got %p (%s)",
			expected, expected.Address,
			got, gotAddr,
		)
		t.Fail()
	}
}