package corn

import (
	"errors"
	"os"
	"strings"

	"github.com/iancoleman/orderedmap"
)

type Value interface {
}

type Evaluation struct {
	Inputs map[string]Rule[any]
	Value  *orderedmap.OrderedMap
}

func evalInputs(assign_block Rule[any], evaluation Evaluation) Evaluation {
	for _, assignment := range assign_block.Rules {
		var nameRule = assignment.Rules[0]
		var name = (*nameRule.Data).(string)

		var value = assignment.Rules[1]

		if value.Id == ruleInput {
			var refName = (*value.Data).(string)
			evaluation.Inputs[name] = evaluation.Inputs[refName]
		} else {
			evaluation.Inputs[name] = value.Rules[0]
		}
	}

	return evaluation
}

func evalValue(val Rule[any], evaluation Evaluation) (Value, error) {
	switch val.Id {
	case ruleObject:
		return evalObject(val, evaluation)
	case ruleArray:
		return evalArray(val, evaluation)
	case ruleBoolean:
		return (*val.Data).(bool), nil
	case ruleFloat:
		return (*val.Data).(float64), nil
	case ruleInteger:
		return (*val.Data).(int64), nil
	case ruleString:
		return evalString(val, evaluation)
	case ruleInput:
		var inputName = (*val.Data).(string)
		return getInput(evaluation, inputName)
	case ruleNull:
		return nil, nil

	default:
		return nil, errors.New("invalid value type: " + val.String())
	}
}

func evalString(rule Rule[any], evaluation Evaluation) (string, error) {
	sb := new(strings.Builder)

	has_escape := false
	for _, rule := range rule.Rules {
		switch rule.Id {
		case ruleCharSequence:
			sb.WriteString((*rule.Data).(string))
		case ruleCharEscape:
			sb.WriteRune((*rule.Data).(rune))
			has_escape = true
		case ruleInput:
			var inputName = (*rule.Data).(string)
			var val, err = getInput(evaluation, inputName)

			if err != nil {
				return "", err
			}

			switch val.(type) {
			case string:
				sb.WriteString(val.(string))
			default:
				return "", errors.New("Attempted to interpolate `" + inputName + "` which is not of type string")
			}
		}
	}

	str := sb.String()

	if !has_escape && strings.Contains(str, "\n") {
		return trimMultilineString(str), nil
	}

	return sb.String(), nil
}

// Takes a multiline string and trims the maximum amount of
// whitespace at the start of each line
// while preserving formatting.
//
// Adapted from corn rust implementation,
// originally based on code from `indoc` crate:
// <https://github.com/dtolnay/indoc/blob/60b5fa29ba4f98b479713621a1f4ec96155caaba/src/unindent.rs#L15-L51>
func trimMultilineString(str string) string {
	ignoreFirstLine := strings.HasPrefix(str, "\n") || strings.HasPrefix(str, "\r\n")

	lines := strings.Split(str, "\n")
	indent := int(^uint(0) >> 1) // init as max

	// first figure out indent depth
	// this should not take the opening line (where the quote is) into consideration
	for _, line := range lines[1:] {
		spaces, _ := takeWhile([]rune(line), func(r rune) bool {
			return r == ' ' || r == '\t'
		})

		// TODO: this +1 may cause problems
		numSpaces := len(spaces) + 1

		if numSpaces < indent {
			indent = numSpaces
		}
	}

	sb := new(strings.Builder)

	// then remove that depth from each line
	for i, line := range lines {
		if i > 1 || (i == 1 && !ignoreFirstLine) {
			sb.WriteRune('\n')
		}

		if i == 0 {
			// do not un-indent anything on same line as opening quote
			sb.WriteString(line)
		} else if len(line) > indent {
			// whitespace-olny lines may have fewer number of spaces being removed
			trimmed := line[indent:]
			sb.WriteString(trimmed)
		}
	}

	return sb.String()
}

func evalArray(arr Rule[any], evaluation Evaluation) ([]Value, error) {
	values := make([]Value, 0, 5)

	for _, rule := range arr.Rules {
		if rule.Id == ruleSpread {
			var inputName = (*rule.Data).(string)
			var value, err = getInput(evaluation, inputName)

			if err != nil {
				return nil, err
			}

			switch value.(type) {
			case []Value:
				values = append(values, value.([]Value)...)
			default:
				return nil, errors.New("attempted to spread non-array input into array")
			}
		} else {

			value, err := evalValue(rule, evaluation)

			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}
	}

	return values, nil
}

func evalObject(obj Rule[any], evaluation Evaluation) (*orderedmap.OrderedMap, error) {
	if obj.Id != ruleObject {
		return orderedmap.New(), errors.New("expected `object`, got " + obj.String())
	}

	value_map := orderedmap.New()

	for _, rule := range obj.Rules {
		if rule.Id == ruleSpread {
			var inputName = (*rule.Data).(string)
			var value, err = getInput(evaluation, inputName)

			if err != nil {
				return value_map, err
			}

			switch value.(type) {
			case map[string]Value:
				for _, k := range value.(*orderedmap.OrderedMap).Keys() {
					v, ok := value.(*orderedmap.OrderedMap).Get(k)

					if !ok {
						return value_map, errors.New("missing key when performing object spread")
					}

					value_map.Set(k, v)
				}
			default:
				return value_map, errors.New("attempted to spread non-object input into object")
			}
		} else {
			var path = evalPath(rule.Rules[0])
			var value, err = evalValue(rule.Rules[1], evaluation)

			if err != nil {
				return value_map, err
			}

			addAtPath(value_map, path, value)
		}
	}

	return value_map, nil
}

func evalPath(path Rule[any]) []string {
	var parts []string

	for _, seg := range path.Rules {
		parts = append(parts, (*seg.Data).(string))
	}

	return parts
}

func addAtPath(obj *orderedmap.OrderedMap, path []string, value Value) error {
	var curr_obj = obj

	for i, seg := range path {
		var is_last = i == len(path)-1

		if is_last {
			curr_obj.Set(seg, value)
			return nil
		}

		child_val, ok := curr_obj.Get(seg)

		if ok {
			var child, ok = child_val.(*orderedmap.OrderedMap)

			if !ok {
				return errors.New("attempted to use key-chaining on non-object type")
			}

			curr_obj = child
		} else {
			child := orderedmap.New()
			curr_obj.Set(seg, child)
			curr_obj = child
		}
	}

	return nil
}

func getInput(evaluation Evaluation, name string) (Value, error) {
	if strings.HasPrefix(name, "$env_") {
		envName := name[len("$env_"):]
		value := os.Getenv(envName)

		if value != "" {
			return value, nil
		}
	}

	rule, ok := evaluation.Inputs[name]

	if ok {
		return evalValue(rule, evaluation)
	} else {
		return nil, errors.New("input '" + name + "' does not exist")
	}
}

func evaluate(ast Rule[any]) (Evaluation, error) {
	var evaluation = Evaluation{
		Inputs: make(map[string]Rule[any]),
	}

	if ast.Id != ruleConfig {
		return evaluation, errors.New("expected `Config`, got " + ast.String())
	}

	var firstRule = ast.Rules[0]
	switch firstRule.Id {
	case ruleAssignBlock:
		evaluation = evalInputs(firstRule, evaluation)

		value, err := evalObject(ast.Rules[1], evaluation)

		evaluation.Value = value

		return evaluation, err

	case ruleObject:
		value, err := evalObject(ast.Rules[0], evaluation)

		evaluation.Value = value

		return evaluation, err

	default:
		return evaluation, errors.New("expected one of `assign_block` or `object`, got " + firstRule.String())
	}
}
