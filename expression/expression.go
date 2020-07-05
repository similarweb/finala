package expression

import (
	"fmt"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

// BoolExpression will parse expression to boolean
func BoolExpression(left interface{}, right interface{}, operator string) (bool, error) {

	formula := fmt.Sprintf("%f %s %f", left.(float64), operator, right.(float64))

	parameters := make(map[string]interface{})
	result, err := ExpressionWithParams(formula, parameters)
	if err != nil {
		return false, err
	}
	return result.(bool), nil
}

// ExpressionWithParams will parse expression
func ExpressionWithParams(formula string, parameters map[string]interface{}) (interface{}, error) {

	expression, err := govaluate.NewEvaluableExpression(formula)

	if err != nil {
		log.WithError(err).WithField("formula", formula).Error("New evaluable expression error")
		return false, err
	}

	log.WithField("formula", formula).Debug("Calculate string expression")

	result, err := expression.Evaluate(parameters)
	if err != nil {
		log.WithError(err).WithField("formula", formula).Error("Evaluation error")
		return false, err
	}

	return result, nil
}
