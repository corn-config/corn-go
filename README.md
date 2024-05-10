# Corn Go

Go implementation of [libcorn](https://github.com/corn-config/corn).

This is a full-spec implementation, which passes the reference test suite, so should work as expected.
That said, it has not currently been road-tested, and the code is rough around the edges, so I would not regard it production-ready.

## Usage


Add the dependency:

```sh
go get github.com/corn-config/corn-go
```

```go
package main

import (
  "fmt"
  "github.com/corn-config/corn-go"
)

func main() {
  input := "let { $foo = 42} in { num = $foo }"
  eval, err := corn.Evaluate(input)

  if err != nil {
    // handle
  }

  fmt.Println(eval.Value.Get("num")) // -> 42
}
```

Support for deserializing into structs is currently missing. 
Only `Evaluate` method is exposed currently, which will return an unstructured map.

