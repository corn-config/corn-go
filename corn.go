package corn

import (
	"errors"
)

// Evaluates the input Corn string, returning an `Evaluation`.
//
// The `Evaluation` consists of two fields -
//
//   - `Inputs` which contains all input values
//   - `Value` which is an ordered map.
//     The map represents the resultant Corn object.
//     Nested objects use nested ordered maps,
//     and arrays are represented using slices.
func Evaluate(input string) (Evaluation, error) {
	tokens := tokenize(input)

	// basic validity checks
	if len(tokens) < 2 {
		return Evaluation{}, errors.New("token stream too short")
	}

	if tokens[0].Id != tokenLet && tokens[0].Id != tokenBraceOpen {
		return Evaluation{}, errors.New("expected first token to be one of `let` or `{`, got" + tokens[0].String())
	}

	if tokens[len(tokens)-1].Id != tokenBraceClose {
		return Evaluation{}, errors.New("expected last token to be `}`, got " + tokens[len(tokens)-1].String())
	}

	ast, err := parse(tokens)

	if err != nil {
		return Evaluation{}, err
	}

	return evaluate(ast)
}
