package webhook

import (
	"testing"

	"k8s.io/apimachinery/pkg/labels"
)

func TestLabelSelector(t *testing.T) {
	selector, err := labels.Parse("bytetrade.io/name=app,app=test")
	if err != nil {
		panic(err)
	}

	labs := labels.Set{
		"app":               "test",
		"bytetrade.io/name": "app",
		"others":            "dev",
	}

	t.Log("matched: ", selector.Matches(labs))
}
