package expression_test

import (
	"testing"

	"finala/expression"
)

func TestExpression(t *testing.T) {

	t.Run("valid_expression", func(t *testing.T) {
		result, err := expression.BoolExpression(float64(1), float64(1), "==")

		if err != nil {
			t.Fatalf("unexpected error response, got %s expected %s", err, "<nil>")
		}

		if !result {
			t.Fatalf("unexpected expression response, got %t expected %t", false, true)
		}

	})

	t.Run("error_expression", func(t *testing.T) {
		_, err := expression.BoolExpression(float64(1), float64(1), "=")
		if err == nil {
			t.Fatalf("unexpected error response, got %s expected %s", "error message", "<nil>")
		}

	})

}
