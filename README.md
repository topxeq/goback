goback
======

goback provides extended regexp syntax, such as Back reference.

The implementation does **NOT** guarantee linear processing time.

## Syntax

See http://godoc.org/github.com/topxeq/goback/regexp/syntax

## Examples

### Back reference

```go
package main

import (
	"fmt"
	"github.com/topxeq/goback/regexp"
)

func main() {
	re := regexp.MustCompile(`^(\w)\w+\k{1}$`)
	fmt.Println(re.MatchString("acca"))       // true
	fmt.Println(re.MatchString("accccab"))    // false
	fmt.Println(re.MatchString("AA"))         // false
}
```

### Possessive qualifiers

```go
re := regexp.MustCompile(`^[0-9]++[0-9a]`)
fmt.Println(re.MatchString("1234a"))     // true
fmt.Println(re.MatchString("1234"))      // false
```

### Atomic group

```go
re := regexp.MustCompile(`^(?>[0-9]+)[0-9a]`)
fmt.Println(re.MatchString("1234a"))     // true
fmt.Println(re.MatchString("1234"))      // false
```

### Comment

```go
re := regexp.MustCompile(`(?#comment here)1234`)
fmt.Println(re.MatchString("1234"))      // true
```

### Lookahead

```go
re := regexp.MustCompile(`a(?=[0-9]{3})1`)
fmt.Println(re.MatchString("a123"))     // true
fmt.Println(re.MatchString("a12a"))     // false
```

### Lookbehind

```go
re := regexp.MustCompile(`(?<=a[0-9]{3,5})a`)
fmt.Println(re.MatchString("a12a"))     // false
fmt.Println(re.MatchString("a12345a"))  // true
```

### Free-Spacing mode

```go
re := regexp.MustCompileFreeSpacing(`

	[0-9]+    # one or more digits
	[a-zA-Z]* # zero or more alphabets
	\#        # literal '#'
	[ ]       # literal ' '

`)
fmt.Println(re.MatchString("1234# "))     // true
fmt.Println(re.MatchString("12345abc "))  // false
```

### Function

```go
re := regexp.MustCompile(`(\d+)\+(\d+)=(?{add})`)

re.Funcs(syntax.FuncMap{
	"add": func(ctx syntax.Context) interface{} {
		lhs, err := strconv.Atoi(string(ctx.Data[ctx.Matches[1][0]:ctx.Matches[1][1]]))
		if err != nil {
			return -1
		}
		rhs, err := strconv.Atoi(string(ctx.Data[ctx.Matches[2][0]:ctx.Matches[2][1]]))
		if err != nil {
			return -1
		}
		answer := strconv.Itoa(lhs + rhs)
		if bytes.HasPrefix(ctx.Data[ctx.Cursor:], []byte(answer)) {
			return len(answer)
		}
		return -1
	},
})

fmt.Println(re.MatchString("12+10=22")) // true
fmt.Println(re.MatchString("1+1=5"))    // false
```
