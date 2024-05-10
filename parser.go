package corn

import (
	"errors"
	"fmt"
)

type ruleId = int

const (
	ruleConfig = iota
	ruleAssignBlock
	ruleObject
	ruleAssignment
	rulePair
	rulePath
	rulePathSegment
	ruleSpread
	ruleArray
	ruleInput
	ruleValue
	ruleBoolean
	ruleFloat
	ruleInteger
	ruleString
	ruleCharSequence
	ruleCharEscape
	ruleNull
)

type Rule[T comparable] struct {
	Id    ruleId
	Rules []Rule[any]
	Data  *T
}

func (r Rule[Stringer]) String() string {
	var identifier string

	switch r.Id {
	case ruleConfig:
		identifier = "Config"
	case ruleAssignBlock:
		identifier = "AssignBlock"
	case ruleObject:
		identifier = "Object"
	case ruleAssignment:
		identifier = "Assignment"
	case rulePair:
		identifier = "Pair"
	case rulePath:
		identifier = "Path"
	case rulePathSegment:
		identifier = "PathSegment"
	case ruleSpread:
		identifier = "Spread"
	case ruleArray:
		identifier = "Array"
	case ruleInput:
		identifier = "Input"
	case ruleValue:
		identifier = "Value"
	case ruleBoolean:
		identifier = "Boolean"
	case ruleFloat:
		identifier = "Float"
	case ruleInteger:
		identifier = "Integer"
	case ruleString:
		identifier = "String"
	case ruleCharSequence:
		identifier = "CharSequence"

	default:
		identifier = "?"
	}

	var string = identifier

	if r.Data != nil {
		string += fmt.Sprintf("(%v)", *r.Data)
	}

	if len(r.Rules) > 0 {
		string += fmt.Sprintf(" %v", r.Rules)
	}

	return string
}

func parseAssignBlock(tokens []Token[any]) (Rule[any], []Token[any], error) {
	var rule = Rule[any]{Id: ruleAssignBlock}

	var brace = tokens[0]
	if brace.Id != tokenBraceOpen {
		return Rule[any]{}, tokens, errors.New("expected `{`, got " + brace.String())
	}

	tokens = tokens[1:]

	var token = tokens[0]
	for token.Id != tokenBraceClose {

		if token.Id != tokenInput {
			return Rule[any]{}, tokens, errors.New("expected `input`, got " + token.String())
		}

		var assignment Rule[any]
		var err error
		assignment, tokens, err = parseAssignment(tokens)
		if err != nil {
			return Rule[any]{}, nil, err
		}

		rule.Rules = append(rule.Rules, assignment)
		token = tokens[0]
	}

	var closeBrace = tokens[0]
	if closeBrace.Id != tokenBraceClose {
		return Rule[any]{}, tokens, errors.New("expected `}`, got " + closeBrace.String())
	}

	var in = tokens[1]
	if in.Id != tokenIn {
		return Rule[any]{}, tokens, errors.New("expected `in`, got " + in.String())
	}

	tokens = tokens[2:]

	return rule, tokens, nil
}

func parseAssignment(tokens []Token[any]) (Rule[any], []Token[any], error) {

	if len(tokens) < 3 {
		return Rule[any]{}, tokens, errors.New("unexpected end of input")
	}

	var input = tokens[0]
	var equals = tokens[1]
	var value = tokens[2]

	if input.Id != tokenInput {
		return Rule[any]{}, tokens, errors.New("expected `input`, got " + input.String())
	}

	if equals.Id != tokenEquals {
		return Rule[any]{}, tokens, errors.New("expected `=`, got " + input.String())
	}

	if value.Id != tokenBraceOpen &&
		value.Id != tokenBracketOpen &&
		value.Id != tokenTrue &&
		value.Id != tokenFalse &&
		value.Id != tokenNull &&
		value.Id != tokenFloat &&
		value.Id != tokenInteger &&
		value.Id != tokenDoubleQuote &&
		value.Id != tokenInput {
		return Rule[any]{}, tokens, errors.New("expected one of `{`, `[`, `true`, `false`, `null`, `float`, `integer`, `\", `input`, got " + input.String())
	}

	valueRule, tokens, err := parseValue(tokens[2:])
	if err != nil {
		return Rule[any]{}, tokens, err
	}

	return Rule[any]{Id: ruleAssignment, Rules: []Rule[any]{
		{Id: ruleInput, Data: input.Data},
		{Id: ruleValue, Rules: []Rule[any]{valueRule}},
	}}, tokens, nil
}

func parseValue(tokens []Token[any]) (Rule[any], []Token[any], error) {
	var token = tokens[0]
	switch token.Id {
	case tokenFloat:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleFloat, Data: token.Data}, tokens, nil
	case tokenInteger:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleInteger, Data: token.Data}, tokens, nil
	case tokenInput:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleInput, Data: token.Data}, tokens, nil
	case tokenTrue:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleBoolean, Data: pointer[any](true)}, tokens, nil
	case tokenFalse:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleBoolean, Data: pointer[any](false)}, tokens, nil
	case tokenDoubleQuote:
		return parseString(tokens)
	case tokenBraceOpen:
		return parseObject(tokens)
	case tokenBracketOpen:
		return parseArray(tokens)
	case tokenNull:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleNull}, tokens, nil
	default:
		return Rule[any]{}, tokens, errors.New("expected `value`, got " + token.String())
	}
}

func pointer[T any](data T) *T {
	return &data
}

func parseObject(tokens []Token[any]) (Rule[any], []Token[any], error) {
	var openBrace = tokens[0]
	if openBrace.Id != tokenBraceOpen {
		return Rule[any]{}, tokens, errors.New("expected `{`, got + " + openBrace.String())
	}

	tokens = tokens[1:]

	var rule = Rule[any]{Id: ruleObject}

	var token = tokens[0]
	for token.Id != tokenBraceClose {
		var pair Rule[any]
		var err error

		switch token.Id {
		case tokenSpread:
			pair, tokens, err = parseSpread(tokens)
		case tokenPathSegment:
			pair, tokens, err = parsePair(tokens)
		default:
			return Rule[any]{}, tokens, errors.New("expected one of `..` or `path_seg`, got " + token.String())
		}

		if err != nil {
			return Rule[any]{}, tokens, err
		}

		rule.Rules = append(rule.Rules, pair)

		token = tokens[0]
	}

	var closeBrace = tokens[0]
	if closeBrace.Id != tokenBraceClose {
		return Rule[any]{}, tokens, errors.New("expected `}`, got " + closeBrace.String())
	}

	tokens = tokens[1:]

	return rule, tokens, nil
}

func parseSpread(tokens []Token[any]) (Rule[any], []Token[any], error) {
	spread := tokens[0]
	if spread.Id != tokenSpread {
		return Rule[any]{}, tokens, errors.New("expected `..`, got " + spread.String())
	}

	input := tokens[1]
	if input.Id != tokenInput {
		return Rule[any]{}, tokens, errors.New("expected `input`, got " + input.String())
	}

	tokens = tokens[2:]

	rule := Rule[any]{Id: ruleSpread, Data: input.Data}
	return rule, tokens, nil
}

func parsePair(tokens []Token[any]) (Rule[any], []Token[any], error) {
	path, tokens, err := parsePath(tokens)

	if err != nil {
		return Rule[any]{}, tokens, err
	}

	var rule = Rule[any]{Id: rulePair, Rules: []Rule[any]{path}}

	var eq = tokens[0]
	if eq.Id != tokenEquals {
		return Rule[any]{}, tokens, errors.New("expected `=`, got " + eq.String())
	}

	tokens = tokens[1:]

	value, tokens, err := parseValue(tokens)

	if err != nil {
		return Rule[any]{}, tokens, err
	}

	rule.Rules = append(rule.Rules, value)

	return rule, tokens, nil
}

func parsePath(tokens []Token[any]) (Rule[any], []Token[any], error) {
	var path_seg = tokens[0]

	if path_seg.Id != tokenPathSegment {
		return Rule[any]{}, tokens, errors.New("expected `path_seg`, got " + path_seg.String())
	}

	var path = Rule[any]{Id: rulePath}

	for path_seg.Id == tokenPathSegment {
		path.Rules = append(path.Rules, Rule[any]{Id: rulePathSegment, Data: path_seg.Data})

		var dot = tokens[1]
		if dot.Id == tokenPathSeparator {
			tokens = tokens[2:]
		} else {
			tokens = tokens[1:]
		}

		path_seg = tokens[0]
	}

	return path, tokens, nil
}

func parseArray(tokens []Token[any]) (Rule[any], []Token[any], error) {
	var openBracket = tokens[0]
	if openBracket.Id != tokenBracketOpen {
		return Rule[any]{}, tokens, errors.New("expected `[`, got " + openBracket.String())
	}

	tokens = tokens[1:]
	var token = tokens[0]

	var rule = Rule[any]{Id: ruleArray}

	var value Rule[any]
	var err error
	for token.Id != tokenBracketClose {
		if token.Id == tokenSpread {
			value, tokens, err = parseSpread(tokens)
		} else {
			value, tokens, err = parseValue(tokens)
		}

		if err != nil {
			return Rule[any]{}, tokens, err
		}

		rule.Rules = append(rule.Rules, value)
		token = tokens[0]
	}

	var closeBracket = tokens[0]
	if closeBracket.Id != tokenBracketClose {
		return Rule[any]{}, tokens, errors.New("expected `]`, got " + closeBracket.String())
	}

	tokens = tokens[1:]

	return rule, tokens, nil
}

func parseString(tokens []Token[any]) (Rule[any], []Token[any], error) {
	var quote = tokens[0]
	if quote.Id != tokenDoubleQuote {
		return Rule[any]{}, tokens, errors.New("expected `\"`, got " + quote.String())
	}

	tokens = tokens[1:]
	var token = tokens[0]

	var rule = Rule[any]{Id: ruleString}

	for token.Id != tokenDoubleQuote {
		switch token.Id {
		case tokenCharSequence:
			rule.Rules = append(rule.Rules, Rule[any]{Id: ruleCharSequence, Data: token.Data})
		case tokenCharEscape:
			rule.Rules = append(rule.Rules, Rule[any]{Id: ruleCharEscape, Data: token.Data})
		case tokenInput:
			rule.Rules = append(rule.Rules, Rule[any]{Id: ruleInput, Data: token.Data})
		default:
			return Rule[any]{}, tokens, errors.New("expected one of `char_seq`, `char_escape` or `input`, got " + token.String())
		}

		tokens = tokens[1:]
		token = tokens[0]
	}

	quote = tokens[0]
	if quote.Id != tokenDoubleQuote {
		return Rule[any]{}, tokens, errors.New("expected `\"`, got " + quote.String())
	}

	tokens = tokens[1:]

	return rule, tokens, nil
}

func parse(tokens []Token[any]) (Rule[any], error) {
	var rule = Rule[any]{Id: ruleConfig}

	var token = tokens[0]
	switch token.Id {
	case tokenLet:
		var assign Rule[any]
		var err error
		assign, tokens, err = parseAssignBlock(tokens[1:])

		if err != nil {
			return rule, err
		}

		rule.Rules = append(rule.Rules, assign)

		var obj Rule[any]
		obj, tokens, err = parseObject(tokens)

		if err != nil {
			return rule, err
		}

		rule.Rules = append(rule.Rules, obj)

	case tokenBraceOpen:
		var obj Rule[any]
		var err error
		obj, tokens, err = parseObject(tokens)

		if err != nil {
			return rule, err
		}

		rule.Rules = append(rule.Rules, obj)

	default:
		return rule, errors.New("expected one of `let` or `{`, got " + token.String())
	}
	return rule, nil
}
