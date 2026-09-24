package trivia_test

import (
	"testing"

	"github.com/AlexS8332/AnimalGuide_Task18/internal/trivia"
	"github.com/AlexS8332/AnimalGuide_Task18/internal/trivia/triviatest"
)

func TestMemoryConformance(t *testing.T) {
	triviatest.PickStoreConformance(t, func(t *testing.T) trivia.PickStore { return trivia.NewMemory() })
}
