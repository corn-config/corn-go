package main

import "fmt"

func main() {
	var input = "let { $bar = 42 } in { foo.bar = $bar baz = false quz = \"bimbo\" pi = 3.14 }"
	var tokens = tokenize(input)

	fmt.Println(tokens)

	var ast, err = parse(tokens)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}

	fmt.Println(ast)

}
