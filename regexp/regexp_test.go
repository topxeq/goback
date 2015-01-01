package regexp

import (
	"fmt"
	"testing"

	gre "regexp"

	"github.com/stretchr/testify/assert"
)

func Example() {
	// Compile the expression once, usually at init time.
	// Use raw strings to avoid having to quote the backslashes.
	var validID = MustCompile(`^(\w)\w+\k{1}\[[0-9]+\]$`)

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

func TestInvalidUTF8(t *testing.T) {
	assert := assert.New(t)
	_, err := compile("\xa7+")
	assert.Error(err, "parse must fail")
}

func TestString(t *testing.T) {
	assert := assert.New(t)
	exp := "[正規表現]"
	assert.Equal(exp, mustCompile(exp).String())
}

func TestSubexp(t *testing.T) {
	assert := assert.New(t)
	exp := "((?P<x>正)(?:現)表現)"

	r := mustCompile(exp)
	g := gre.MustCompile(exp)
	assert.Equal(g.NumSubexp(), r.NumSubexp())
	assert.Equal(g.SubexpNames(), r.SubexpNames())
}

func TestLongest(t *testing.T) {
	assert := assert.New(t)
	exp := "[正規表現]{0,5}?"
	str := "正規表現正規表現"

	r := mustCompile(exp)
	r.Longest()

	g := gre.MustCompile(exp)
	g.Longest()

	assert.Equal(g.FindStringSubmatchIndex(str), r.FindStringSubmatchIndex(str))
}

func TestLiteral(t *testing.T) {
	assert := assert.New(t)
	exp := "\\Q...\\E.+"
	str := "正規表現...正規表現"

	r := mustCompile(exp)
	g := gre.MustCompile(exp)

	assert.Equal(g.FindStringSubmatchIndex(str), r.FindStringSubmatchIndex(str))
}

func TestSplit(t *testing.T) {
	assert := assert.New(t)
	{
		exp := "[正現]"
		str := "regex正規表現正規表現正規表現regex"

		r := mustCompile(exp)
		g := gre.MustCompile(exp)

		for i := -1; i < 5; i++ {
			assert.Equal(g.Split(str, i), r.Split(str, i))
		}
	}
	{
		exp := "せ?"
		str := "regex正規表現正規表現正規表現regex"

		r := mustCompile(exp)
		g := gre.MustCompile(exp)

		for i := -1; i < 5; i++ {
			assert.Equal(g.Split(str, i), r.Split(str, i))
		}
	}
}

func TestExpand(t *testing.T) {
	assert := assert.New(t)
	exp := "(?P<reg>([正現]*)規)"
	tmp := "${reg}-$1-$0-$reg"
	str := "正規表現正規表現正規表現"

	r := mustCompile(exp)
	g := gre.MustCompile(exp)

	assert.Equal(g.ExpandString(nil, tmp, str, g.FindStringSubmatchIndex(str)),
		r.ExpandString(nil, tmp, str, r.FindStringSubmatchIndex(str)))
}

func TestReplace(t *testing.T) {
	assert := assert.New(t)
	exp := "(?P<reg>([正現]*)規)"
	tmp := "${reg}-$1-$0-$reg"
	str := "正規表現正規表現正規表現"

	r := mustCompile(exp)
	g := gre.MustCompile(exp)

	assert.Equal(g.ReplaceAllString(str, tmp), r.ReplaceAllString(str, tmp))
	assert.Equal(g.ReplaceAllLiteralString(str, tmp), r.ReplaceAllLiteralString(str, tmp))
}

func TestRepeat(t *testing.T) {
	assert := assert.New(t)

	str := "正規表現 正規 表現 せいきひょうげん regexp"
	exp := []string{
		"[正規表現]{3}",
		"[正規表現]{3,5}",
		"[正規表現]{6,}?",
		"[正規表現]{3}?",
		"[正規表現]{3,5}?",
		"[正規表現]?",
		"[正規表現]*",
		"[正規表現]+",
		"[正規表現]??",
		"[正規表現]*?",
		"[正規表現]+?",
	}

	for _, e := range exp {
		r := mustCompile(e)
		g := gre.MustCompile(e)
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}

	for _, e := range exp {
		r := mustCompile(e)
		r.Longest()
		g := gre.MustCompile(e)
		g.Longest()
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}

	errexp := []string{
		"[正規表現]{3000}",
		"[正規表現]{3,5000}",
		"[正規表現]**",
		"[正規表現]+*",
		"[正規表現]???",
	}

	for _, e := range errexp {
		_, err := compile(e)
		assert.Error(err, "parse must fail")
	}
}

func TestAlternation(t *testing.T) {
	assert := assert.New(t)

	str := "正規表現 正規 表現 せいきひょうげん regexp"
	exp := []string{
		"正規表現|正規|表現|",
		"|||正規表現||||正規|| ||表現||||",
		"正規*?表現|正??| |表??現",
	}

	for _, e := range exp {
		r := mustCompile(e)
		g := gre.MustCompile(e)
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}

	for _, e := range exp {
		r := mustCompile(e)
		r.Longest()
		g := gre.MustCompile(e)
		g.Longest()
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}
}

func TestBackref(t *testing.T) {
	assert := assert.New(t)

	str := "正規表現 正規 表現 せいきひょうげん regexp"
	exp := "(?P<f>正規)?(?P<s>表現)? \\k1 \\k{s} \\k2*"
	r := mustCompile(exp)
	assert.Equal([][]int{[]int{0, 27, 0, 6, 6, 12}},
		r.FindAllStringSubmatchIndex(str, -1))
}

func TestFlags(t *testing.T) {
	assert := assert.New(t)

	str := "\n正規表現 正規 表現\n せいきひょうげん regexp\n"
	exp := []string{
		"(?U)[正規表現]?",
		"(?U)[正規表現]*",
		"(?U)(?-U)[正規表現]+",
		"(?U)[正規表現]??",
		"(?U)[正規表現]*?",
		"(?U)(?-U)[正規表現]+?",
		".+",
		"(?s).+",
		"(?i)REGEXP",
		"(?i)REG(?-i)EXP",
		"(?i:REGEXP)",
		"(?i:REG(?-i)EXP)",
		"^.+$",
		"(?m)^.*$",
		"(?m:^^.*?.*$$)",
		"(?m)^^(?-m).*$$",
		"(?m)\\A.*\\z",
		"(?m:\\A\\A.*\\z\\z)",
	}

	for _, e := range exp {
		r := mustCompile(e)
		g := gre.MustCompile(e)
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}
}

func TestCharClass(t *testing.T) {
	assert := assert.New(t)

	str := "\n正規表現 正規 表現\n せいきひょうげん 325-6204 regexp\\\\ Regexp\n"
	exp := []string{
		"[せ-ん正-現]?",
		"[^\n]+",
		"\\d+",
		"\\D+",
		"\\s+",
		"\\S+",
		"\\w+",
		"\\W+",
		"[[]",
		"[]]",
		"[\\-]",
		"[\\d]+",
		"[\\D]+",
		"[\\s]+",
		"[\\S]+",
		"[\\w]+",
		"[\\W]+",
		"[^\\d]+",
		"[^\\D]+",
		"[^\\s]+",
		"[^\\S]+",
		"[^\\w]+",
		"[^\\W]+",
		"[:space\\:]+",
		"[[:alnum:]]+",
		"[[:^alnum:]]+",
		"[^[:alnum:]]+",
		"[^[:^alnum:]]+",
		"[[:alpha:]]+",
		"[[:^alpha:]]+",
		"[^[:alpha:]]+",
		"[^[:^alpha:]]+",
		"[[:ascii:]]+",
		"[[:^ascii:]]+",
		"[^[:ascii:]]+",
		"[^[:^ascii:]]+",
		"[[:blank:]]+",
		"[[:^blank:]]+",
		"[^[:blank:]]+",
		"[^[:^blank:]]+",
		"[[:cntrl:]]+",
		"[[:^cntrl:]]+",
		"[^[:cntrl:]]+",
		"[^[:^cntrl:]]+",
		"[[:graph:]]+",
		"[[:^graph:]]+",
		"[^[:graph:]]+",
		"[^[:^graph:]]+",
		"[[:lower:]]+",
		"[[:^lower:]]+",
		"[^[:lower:]]+",
		"[^[:^lower:]]+",
		"[[:print:]]+",
		"[[:^print:]]+",
		"[^[:print:]]+",
		"[^[:^print:]]+",
		"[[:punct:]]+",
		"[[:^punct:]]+",
		"[^[:punct:]]+",
		"[^[:^punct:]]+",
		"[[:upper:]]+",
		"[[:^upper:]]+",
		"[^[:upper:]]+",
		"[^[:^upper:]]+",
		"[[:xdigit:]]+",
		"[[:^xdigit:]]+",
		"[^[:xdigit:]]+",
		"[^[:^xdigit:]]+",
		"[[:digit:]]+",
		"[[:^digit:]]+",
		"[^[:digit:]]+",
		"[^[:^digit:]]+",
		"[[:word:]]+",
		"[[:^word:]]+",
		"[^[:word:]]+",
		"[^[:^word:]]+",
		"[\\p{Hiragana}]+",
		"[\\P{Hiragana}]+",
		"[\\pC]+",
		"[\\PC]+",
	}

	for _, e := range exp {
		r := mustCompile(e)
		g := gre.MustCompile(e)
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}

	errexp := []string{
		"[z-a]",
		"[-現-正]",
	}

	for _, e := range errexp {
		_, err := compile(e)
		assert.Error(err, "parse must fail")
	}
}

func TestEscape(t *testing.T) {
	assert := assert.New(t)

	str := "\n正規表現 正規 表現\n せいきひょうげん 325-6204 regexp\n \a\f\t\n\r\v"
	exp := []string{
		"\\a?\\f?\\t?\\n?\\r?\\v?",
		"[\\a\\f\\t\\n\\r\\v]+",
		"\\b",
		"\\B",
		"\\p{Hiragana}",
		"\\P{Hiragana}",
		"\\pC",
		"\\PC",
		"\\123",
		"\\23",
		"\\032",
		"\\x20",
		"\\x{20}",
		"\\x{1F}",
	}

	for _, e := range exp {
		r := mustCompile(e)
		g := gre.MustCompile(e)
		assert.Equal(g.FindAllStringSubmatchIndex(str, -1),
			r.FindAllStringSubmatchIndex(str, -1))
	}
}
