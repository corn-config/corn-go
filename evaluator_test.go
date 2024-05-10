package corn

import (
	"encoding/json"
	"fmt"
	"github.com/andreyvit/diff"
	"os"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, actual string, expected string) {
	if expected != actual {
		msg := "expected and actual are not equal\n\n"
		t.Fatalf(msg + diff.LineDiff(expected, actual))
		// t.Fatalf(msg)
	}
}

func testEqual(t *testing.T, testName string) {
	test_path := "../corn/assets"

	input := test_path + "/inputs/" + testName + ".corn"
	output := test_path + "/outputs/json/" + testName + ".json"

	inputBytes, err := os.ReadFile(input)

	if err != nil {
		fmt.Println("ERROR ", err)
		return
	}

	inputStr := string(inputBytes)

	outputBytes, err := os.ReadFile(output)

	if err != nil {
		t.Skipf(err.Error())
		return
	}

	outputStr := strings.TrimSpace(string(outputBytes))

	evaluation, err := Evaluate(inputStr)

	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	inputJson, err := json.MarshalIndent(evaluation.Value, "", "  ")

	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	assertEqual(t, string(inputJson), outputStr)

}

func TestArray(t *testing.T) {
	testEqual(t, "array")
}

func TestBasic(t *testing.T) {
	testEqual(t, "basic")
}

func TestBasicEmptyLet(t *testing.T) {
	testEqual(t, "basic_empty_let")
}

func TestBoolean(t *testing.T) {
	testEqual(t, "boolean")
}

func TestChained(t *testing.T) {
	testEqual(t, "chained")
}

func TestChainedComplex(t *testing.T) {
	testEqual(t, "chained_complex")
}

func TestChar(t *testing.T) {
	testEqual(t, "char")
}

func TestComment(t *testing.T) {
	testEqual(t, "comment")
}

func TestCompact(t *testing.T) {
	testEqual(t, "compact")
}

func TestComplex(t *testing.T) {
	testEqual(t, "complex")
}

// FIXME: Discrepancy in JSON behaviour? Also looks like weird result in Rust
func TestComplexKeys(t *testing.T) {
	testEqual(t, "complex_keys")
}

func TestEnvironmentVariable(t *testing.T) {
	testEqual(t, "environment_variable")
}

func TestFloat(t *testing.T) {
	testEqual(t, "float")
}

func TestInput(t *testing.T) {
	testEqual(t, "input")
}

func TestInputReferencesInput(t *testing.T) {
	testEqual(t, "input_references_input")
}

func TestInteger(t *testing.T) {
	testEqual(t, "integer")
}

func TestInvalid(t *testing.T) {
	// TODO: Write test
}

func TestInvalidInput(t *testing.T) {
	// TODO: write test
}

func TestInvalidNesting(t *testing.T) {
	// TODO: write test
}

func TestInvalidSpread(t *testing.T) {
	// TODO: Write test
}

func TestMixedArray(t *testing.T) {
	testEqual(t, "mixed_array")
}

func TestNull(t *testing.T) {
	testEqual(t, "null")
}

func TestObject(t *testing.T) {
	testEqual(t, "object")
}

func TestObjectInArray(t *testing.T) {
	testEqual(t, "object_in_array")
}

func TestQuotedKeys(t *testing.T) {
	testEqual(t, "quoted_keys")
}

func TestReadmeExample(t *testing.T) {
	testEqual(t, "readme_example")
}

func TestSpread(t *testing.T) {
	testEqual(t, "spread")
}

func TestString(t *testing.T) {
	testEqual(t, "string")
}

func TestStringMultiline(t *testing.T) {
	testEqual(t, "string_multiline")
}

func TestStringInterpolation(t *testing.T) {
	testEqual(t, "string_interpolation")
}

func TestValueAfterTable(t *testing.T) {
	testEqual(t, "value_after_table")
}

func TestVeryCompact(t *testing.T) {
	testEqual(t, "very_compact")
}
