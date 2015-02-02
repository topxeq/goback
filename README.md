goback
======

[![Build Status](https://travis-ci.org/h2so5/goback.svg)](https://travis-ci.org/h2so5/goback)
[![GoDoc](https://godoc.org/github.com/h2so5/goback/regexp?status.svg)](http://godoc.org/github.com/h2so5/goback/regexp)

goback provides extended regexp syntax, such as Back reference.

The implementation does **NOT** guarantee linear processing time.

## Examples

### Back reference

```go
package main

import (
	"fmt"
	"github.com/h2so5/goback/regexp"
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
