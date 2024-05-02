package main

import (
	"fmt"
	"strconv"
	"strings"
)

type tokenId = int32
type stateId = int8

const (
	tokenBraceOpen = iota
	tokenBraceClose
	tokenBracketOpen
	tokenBracketClose
	tokenEquals
	tokenDoubleQuote
	tokenSpread
	tokenPathSeparator
	tokenLet
	tokenIn
	tokenTrue
	tokenFalse
	tokenNull
	tokenPathSegment
	tokenFloat
	tokenInteger
	tokenCharSequence
	tokenInput
)

const (
	charsWhitespace          = " \t\r\n"
	charsInvalidPath         = charsWhitespace + "=."
	charsInteger             = "0123456789"
	charsInputFirst          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsInput               = charsInputFirst + "1234567890_"
	charsInvalidCharSequence = "\"\\$"
)

const (
	stateTopLevel = iota
	stateAssignBlock
	stateValue
	stateObject
	stateArray
	stateString
)

type matcher = func(input []rune, tokens []Token[any]) ([]rune, []Token[any], bool)
type dynamicTokenMatcher = func(input []rune) ([]rune, any)
type stateChanger = func(state []stateId) []stateId

type MatchRule struct {
	matcher     matcher
	stateChange stateChanger
}

func matchRule(matcher matcher) MatchRule {
	return MatchRule{matcher: matcher}
}

func matchRuleStateChange(matcher matcher, stateChange stateChanger) MatchRule {
	return MatchRule{
		matcher:     matcher,
		stateChange: stateChange,
	}
}

type Token[T comparable] struct {
	Id   tokenId
	Data *T
}

func (t Token[Stringer]) String() string {
	var identifier string
	switch t.Id {
	case tokenBraceOpen:
		identifier = "{"
	case tokenBraceClose:
		identifier = "}"
	case tokenBracketOpen:
		identifier = "["
	case tokenEquals:
		identifier = "="
	case tokenDoubleQuote:
		identifier = "\""
	case tokenSpread:
		identifier = ".."
	case tokenPathSeparator:
		identifier = "."
	case tokenLet:
		identifier = "let"
	case tokenIn:
		identifier = "in"
	case tokenTrue:
		identifier = "true"
	case tokenFalse:
		identifier = "false"
	case tokenNull:
		identifier = "null"
	case tokenPathSegment:
		identifier = "path_seg"
	case tokenFloat:
		identifier = "float"
	case tokenInteger:
		identifier = "int"
	case tokenInput:
		identifier = "input"
	case tokenCharSequence:
		identifier = "char_seq"

	default:
		identifier = "?"
	}

	if t.Data != nil {
		return fmt.Sprintf("Token(%s(%v))", identifier, *t.Data)

	} else {
		return fmt.Sprintf("Token(%s)", identifier)
	}
}

func simpleToken(id tokenId) Token[any] {
	return Token[any]{Id: id}
}

func dataToken[T comparable](id tokenId, data T) Token[T] {
	return Token[T]{Id: id, Data: &data}
}

func takeWhile(input []rune, check func(rune) bool) ([]rune, []rune) {
	var slice []rune

	var char = input[0]
	for check(char) && len(input) > 1 {
		var char2, input2 = input[0], input[1:]
		char = char2

		// TODO: Two checks - bad :(
		if check(char) {
			input = input2
			slice = append(slice, char)
		}
	}

	return slice, input
}

func runesEqual(a []rune, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func matchToken(tokenId tokenId, tokenChar rune) matcher {
	return func(input []rune, tokens []Token[any]) ([]rune, []Token[any], bool) {
		var match = false
		if len(input) > 0 && input[0] == tokenChar {
			input = input[1:]
			tokens = append(tokens, simpleToken(tokenId))
			match = true
		}

		return input, tokens, match
	}
}

func matchCompoundToken(tokenId tokenId, tokenChars []rune) matcher {
	return func(input []rune, tokens []Token[any]) ([]rune, []Token[any], bool) {
		if len(input) < len(tokenChars) {
			return input, tokens, false
		}

		if runesEqual(input[:len(tokenChars)], tokenChars) {
			input = input[len(tokenChars):]
			tokens = append(tokens, simpleToken(tokenId))

			return input, tokens, true
		}

		return input, tokens, false
	}
}

func matchDynamicToken(tokenId tokenId, matcher dynamicTokenMatcher) matcher {
	return func(input []rune, tokens []Token[any]) ([]rune, []Token[any], bool) {
		var data any
		input, data = matcher(input)

		if data != nil {
			var token = dataToken[any](tokenId, data)
			tokens = append(tokens, token)
			return input, tokens, true
		}

		return input, tokens, false
	}
}

func matchInput(input []rune) ([]rune, any) {
	if len(input) < 2 || input[0] != '$' {
		return input, nil
	}

	var name = []rune{input[0]}

	if strings.ContainsRune(charsInputFirst, input[1]) {
		name = append(name, input[1])

		input = input[2:]

		var rest []rune
		rest, input = takeWhile(input, func(char rune) bool {
			return strings.ContainsRune(charsInput, char)
		})

		name = append(name, rest...)
		return input, string(name)
	}

	return input, nil
}

func matchPathSegment(input []rune) ([]rune, any) {
	var path []rune
	path, input = takeWhile(input, func(r rune) bool {
		return !strings.ContainsRune(charsInvalidPath, r)
	})

	if len(path) > 0 {
		return input, string(path)
	}

	return input, nil
}

func matchFloat(input []rune) ([]rune, any) {
	var numS []rune
	numS, _ = takeWhile(input, func(char rune) bool {
		return strings.ContainsRune(charsInteger, char)
	})

	if len(input) < len(numS)+2 || input[len(numS)] != '.' {
		return input, nil
	}

	input = input[len(numS)+1:]

	var decimals []rune
	decimals, input = takeWhile(input, func(char rune) bool {
		return strings.ContainsRune(charsInteger, char)
	})

	numS = append(numS, '.')
	numS = append(numS, decimals...)
	var num, err = strconv.ParseFloat(string(numS), 64)

	if err != nil {
		return input, nil
	}

	return input, num
}

func matchInteger(input []rune) ([]rune, any) {
	var numS []rune
	numS, input = takeWhile(input, func(char rune) bool {
		return strings.ContainsRune(charsInteger, char)
	})

	var num, err = strconv.Atoi(string(numS))
	if err != nil {
		return input, nil
	}

	return input, num
}

func matchCharSequence(input []rune) ([]rune, any) {
	var seq []rune
	seq, input = takeWhile(input, func(char rune) bool {
		// TODO: Handle escape chars
		return !strings.ContainsRune(charsInvalidCharSequence, char)
	})

	if len(seq) > 0 {
		return input, string(seq)
	}

	return input, nil
}

// TODO: Make more efficient - allocating on every token currently :(
func getMatchers(state stateId) []MatchRule {
	switch state {
	case stateTopLevel:
		return []MatchRule{
			matchRuleStateChange(matchCompoundToken(tokenLet, []rune("let")), statePusher(stateAssignBlock)),
			matchRuleStateChange(matchToken(tokenBraceOpen, '{'), statePusher(stateObject)),
		}
	case stateAssignBlock:
		return []MatchRule{
			matchRuleStateChange(matchCompoundToken(tokenIn, []rune("in")), popState),
			matchRule(matchToken(tokenBraceOpen, '{')),
			matchRule(matchToken(tokenBraceClose, '}')),
			matchRule(matchDynamicToken(tokenInput, matchInput)),
			matchRuleStateChange(matchToken(tokenEquals, '='), statePusher(stateValue)),
		}
	case stateObject:
		return []MatchRule{
			matchRuleStateChange(matchToken(tokenBraceClose, '}'), popState),

			matchRule(matchDynamicToken(tokenPathSegment, matchPathSegment)),
			matchRuleStateChange(matchToken(tokenEquals, '='), statePusher(stateValue)),
			matchRule(matchCompoundToken(tokenSpread, []rune(".."))),
			matchRule(matchToken(tokenPathSeparator, '.')),
			matchRule(matchDynamicToken(tokenInput, matchInput)),
		}
	case stateArray:
		return []MatchRule{
			matchRuleStateChange(matchToken(tokenBracketClose, ']'), popState),

			matchRuleStateChange(matchToken(tokenBraceOpen, '{'), statePusher(stateObject)),
			matchRuleStateChange(matchToken(tokenBracketOpen, '['), statePusher(stateArray)),

			matchRule(matchCompoundToken(tokenSpread, []rune(".."))),

			matchRule(matchCompoundToken(tokenTrue, []rune("true"))),
			matchRule(matchCompoundToken(tokenFalse, []rune("false"))),
			matchRule(matchCompoundToken(tokenNull, []rune("null"))),

			matchRuleStateChange(matchToken(tokenDoubleQuote, '"'), statePusher(stateString)),

			matchRule(matchDynamicToken(tokenInput, matchInput)),
			matchRule(matchDynamicToken(tokenFloat, matchFloat)),
			matchRule(matchDynamicToken(tokenInteger, matchInteger)),
		}
	case stateValue:
		return []MatchRule{
			matchRuleStateChange(matchToken(tokenBraceOpen, '{'), stateReplacer(stateObject)),
			matchRuleStateChange(matchToken(tokenBracketOpen, '['), stateReplacer(stateArray)),

			matchRuleStateChange(matchCompoundToken(tokenTrue, []rune("true")), popState),
			matchRuleStateChange(matchCompoundToken(tokenFalse, []rune("false")), popState),
			matchRuleStateChange(matchCompoundToken(tokenNull, []rune("null")), popState),

			matchRuleStateChange(matchToken(tokenDoubleQuote, '"'), stateReplacer(stateString)),

			matchRuleStateChange(matchDynamicToken(tokenInput, matchInput), popState),
			matchRuleStateChange(matchDynamicToken(tokenFloat, matchFloat), popState),
			matchRuleStateChange(matchDynamicToken(tokenInteger, matchInteger), popState),

			matchRuleStateChange(matchToken(tokenBracketClose, ']'), popState),
		}
	case stateString:
		return []MatchRule{
			matchRuleStateChange(matchToken(tokenDoubleQuote, '"'), popState),

			matchRule(matchDynamicToken(tokenInput, matchInput)),
			matchRule(matchDynamicToken(tokenCharSequence, matchCharSequence)),
		}
	default:
		panic("invalid state id " + strconv.Itoa(int(state)))
	}
}

func statePusher(newState stateId) stateChanger {
	return func(state []stateId) []stateId {
		return append(state, newState)
	}
}

func stateReplacer(newState stateId) stateChanger {
	return func(state []stateId) []stateId {
		return append(state[:len(state)-1], newState)
	}
}

func popState(state []stateId) []stateId {
	return state[:len(state)-1]
}

func tokenize(inputString string) []Token[any] {
	var input = []rune(inputString)
	var tokens []Token[any]
	var state = []stateId{stateTopLevel}

	for len(input) > 0 {
		var length = len(input)

		//fmt.Println(string(input))
		//fmt.Println(tokens)
		//fmt.Println("---")

		if strings.ContainsRune(charsWhitespace, input[0]) {
			_, input = takeWhile(input, func(r rune) bool {
				return strings.ContainsRune(charsWhitespace, r)
			})
		}

		var currentState = state[len(state)-1]

		var matchers = getMatchers(currentState)

		for _, matcher := range matchers {
			var match bool
			input, tokens, match = matcher.matcher(input, tokens)

			if match {
				if matcher.stateChange != nil {
					state = matcher.stateChange(state)
				}

				break
			}
		}

		//fmt.Println(string(input))
		//fmt.Println(tokens)
		//fmt.Println("=====")

		if len(input) == length {
			panic("length unchanged")
		}
	}

	return tokens
}
