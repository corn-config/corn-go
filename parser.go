package main

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
	ruleInput
	ruleValue
	ruleInteger
)

type Rule[T comparable] struct {
	Id    ruleId
	Rules []Rule[any]
	Data  *T
}

//func getValidTokens(ruleId ruleId) []tokenId {
//	switch ruleId {
//	case ruleConfig:
//		return []tokenId{tokenLet, tokenBraceOpen}
//	case ruleAssignBlock:
//		return []tokenId {}
//	case ruleObject:
//		return []tokenId{tokenPathSegment}
//	default:
//		panic("unimplemented rule")
//	}
//}
//
//func getRuleForToken(tokenId tokenId) Rule {
//	switch tokenId {
//	case tokenLet
//	}
//}

func parseAssignBlock(tokens []Token[any]) (Rule[any], error) {
	fmt.Println("parse assign block", tokens)

	var rule = Rule[any]{Id: ruleAssignBlock}

	var brace = tokens[0]
	if brace.Id != tokenBraceOpen {
		return Rule[any]{}, errors.New("expected `{`, got " + brace.String())
	}

	tokens = tokens[1:]

	var token = tokens[0]
	for token.Id != tokenBraceClose {
		fmt.Println(token)

		if token.Id != tokenInput {
			return Rule[any]{}, errors.New("expected `input`, got " + token.String())
		}

		var assignment, err = parseAssignment(tokens)
		if err != nil {
			return Rule[any]{}, err
		}

		rule.Rules = append(rule.Rules, assignment)

		// tokens = tokens[1:]
		token = tokens[0]
	}

	return rule, nil
}

func parseAssignment(tokens []Token[any]) (Rule[any], error) {
	fmt.Println("parse assignment", tokens)

	if len(tokens) < 3 {
		return Rule[any]{}, errors.New("unexpected end of input")
	}

	var input = tokens[0]
	var equals = tokens[1]
	var value = tokens[2]

	if input.Id != tokenInput {
		return Rule[any]{}, errors.New("expected `input`, got " + input.String())
	}

	if equals.Id != tokenEquals {
		return Rule[any]{}, errors.New("expected `=`, got " + input.String())
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
		return Rule[any]{}, errors.New("expected one of `{`, `[`, `true`, `false`, `null`, `float`, `integer`, `\", `input`, got " + input.String())
	}

	var valueRule, err = parseValue(tokens[2:])
	if err != nil {
		return Rule[any]{}, err
	}

	return Rule[any]{Id: ruleAssignment, Rules: []Rule[any]{
		{Id: ruleInput, Data: input.Data},
		{Id: ruleValue, Rules: []Rule[any]{valueRule}},
	}}, nil
}

func parseValue(tokens []Token[any]) (Rule[any], error) {
	var token = tokens[0]
	switch token.Id {
	case tokenInteger:
		tokens = tokens[1:]
		return Rule[any]{Id: ruleInteger, Data: token.Data}, nil
	default:
		return Rule[any]{}, errors.New("expected value, got " + token.String())
	}
}

func parseObject(tokens []Token[any]) {
	fmt.Println("parse object")
}

func parseArray(tokens []Token[any]) {

}

func contains(slice []tokenId, tokenId tokenId) bool {
	for _, id := range slice {
		if id == tokenId {
			return true
		}
	}

	return false
}

func parse(tokens []Token[any]) (Rule[any], error) {
	var rule = Rule[any]{Id: ruleConfig}

	var token = tokens[0]
	switch token.Id {
	case tokenLet:
		var assign, err = parseAssignBlock(tokens[1:])
		if err != nil {
			return rule, err
		}
		rule.Rules = append(rule.Rules, assign)
	case tokenBraceOpen:
		parseObject(tokens)
	default:
		return rule, errors.New("expected one of `let` or `{`, got " + token.String())
	}

	//var currRule = rule
	//
	//for _, token := range tokens {
	//	var validTokens = getValidTokens(currRule.Id)
	//
	//	fmt.Println("valid ", validTokens)
	//
	//	if contains(validTokens, token.Id) {
	//		currRule.Rules
	//	}
	//
	//	fmt.Println(token)
	//}

	return rule, nil
}
