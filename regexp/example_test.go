package regexp_test

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/Upliner/goback/regexp"
	"github.com/Upliner/goback/regexp/syntax"
)

func Example() {
	// Compile the expression once, usually at init time.
	// Use raw strings to avoid having to quote the backslashes.
	var validID = regexp.MustCompile(`^(\w)\w+\k{1}\[[0-9]+\]$`)

	fmt.Println(validID.MatchString("adam[23]"))
	fmt.Println(validID.MatchString("eve[7]"))
	fmt.Println(validID.MatchString("Job[48]"))
	fmt.Println(validID.MatchString("snakey"))
	// Output:
	// false
	// true
	// false
	// false
}

func Example_BackReference() {
	re := regexp.MustCompile(`^(\w)\w+\k{1}$`)
	fmt.Println(re.MatchString("acca"))
	fmt.Println(re.MatchString("accccab"))
	fmt.Println(re.MatchString("AA"))
	// Output:
	// true
	// false
	// false
}

func Example_PossessiveQualifiers() {
	re := regexp.MustCompile(`^[0-9]++[0-9a]`)
	fmt.Println(re.MatchString("1234a"))
	fmt.Println(re.MatchString("1234"))
	// Output:
	// true
	// false
}

func Example_AtomicGroup() {
	re := regexp.MustCompile(`^(?>[0-9]+)[0-9a]`)
	fmt.Println(re.MatchString("1234a"))
	fmt.Println(re.MatchString("1234"))
	// Output:
	// true
	// false
}

func Example_Comment() {
	re := regexp.MustCompile(`(?#comment here)1234`)
	fmt.Println(re.MatchString("1234"))
	// Output:
	// true
}

func Example_Lookahead() {
	re := regexp.MustCompile(`a(?=[0-9]{3})1`)
	fmt.Println(re.MatchString("a123"))
	fmt.Println(re.MatchString("a12a"))
	// Output:
	// true
	// false
}

func Example_Lookbehind() {
	re := regexp.MustCompile(`(?<=a[0-9]{3,5})a`)
	fmt.Println(re.MatchString("a12a"))
	fmt.Println(re.MatchString("a12345a"))
	// Output:
	// false
	// true
}

func Example_FreeSpacing() {
	re := regexp.MustCompileFreeSpacing(`

		[0-9]+    # one or more digits
		[a-zA-Z]* # zero or more alphabets
		\#        # literal '#'
		[ ]       # literal ' '

	`)
	fmt.Println(re.MatchString("1234# "))
	fmt.Println(re.MatchString("12345abc "))
	// Output:
	// true
	// false
}

func Example_Function() {
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
	fmt.Println(re.MatchString("12+10=22"))
	fmt.Println(re.MatchString("1+1=5"))
	// Output:
	// true
	// false
}
