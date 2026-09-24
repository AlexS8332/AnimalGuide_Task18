package mdd_test

import (
	"testing"

	"github.com/AlexS8332/AnimalGuide_Task18/internal/mdd"
	"github.com/AlexS8332/AnimalGuide_Task18/internal/mdd/mddtest"
)

func TestMemoryConformance(t *testing.T) {
	mddtest.Conformance(t, func(t *testing.T) mdd.Store { return mdd.NewMemory() })
}
