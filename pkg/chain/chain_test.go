package chain_test

import (
	"os"
	"testing"

	"github.com/yourorg/envchain/pkg/chain"
)

func TestResolve_HigherPriorityWins(t *testing.T) {
	c := chain.New()
	c.AddSource(&chain.Source{Name: "low", Priority: 1, Vars: map[string]string{"KEY": "low-value"}})
	c.AddSource(&chain.Source{Name: "high", Priority: 10, Vars: map[string]string{"KEY": "high-value"}})

	val, ok := c.Resolve("KEY")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "high-value" {
		t.Errorf("expected %q, got %q", "high-value", val)
	}
}

func TestResolve_FallbackToEnv(t *testing.T) {
	os.Setenv("FALLBACK_KEY", "from-env")
	t.Cleanup(func() { os.Unsetenv("FALLBACK_KEY") })

	c := chain.New()
	val, ok := c.Resolve("FALLBACK_KEY")
	if !ok {
		t.Fatal("expected fallback to process env")
	}
	if val != "from-env" {
		t.Errorf("expected %q, got %q", "from-env", val)
	}
}

func TestResolve_MissingKey(t *testing.T) {
	c := chain.New()
	_, ok := c.Resolve("DOES_NOT_EXIST_XYZ")
	if ok {
		t.Error("expected key to be missing")
	}
}

func TestResolveAll_MergesAllSources(t *testing.T) {
	c := chain.New()
	c.AddSource(&chain.Source{Name: "a", Priority: 1, Vars: map[string]string{"A": "1", "SHARED": "a-val"}})
	c.AddSource(&chain.Source{Name: "b", Priority: 5, Vars: map[string]string{"B": "2", "SHARED": "b-val"}})

	all := c.ResolveAll()

	if all["A"] != "1" {
		t.Errorf("A: expected %q got %q", "1", all["A"])
	}
	if all["B"] != "2" {
		t.Errorf("B: expected %q got %q", "2", all["B"])
	}
	if all["SHARED"] != "b-val" {
		t.Errorf("SHARED: expected %q got %q", "b-val", all["SHARED"])
	}
}

func TestInject_SetsProcessEnv(t *testing.T) {
	t.Cleanup(func() { os.Unsetenv("INJECT_TEST") })

	c := chain.New()
	c.AddSource(&chain.Source{Name: "src", Priority: 1, Vars: map[string]string{"INJECT_TEST": "injected"}})

	if err := c.Inject(); err != nil {
		t.Fatalf("Inject returned error: %v", err)
	}
	if got := os.Getenv("INJECT_TEST"); got != "injected" {
		t.Errorf("expected %q, got %q", "injected", got)
	}
}
