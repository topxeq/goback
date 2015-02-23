package regexp

import (
	"bufio"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"reflect"
	gre "regexp"
)

func AssertBuiltIn(t *testing.T, exp, str string) {
	r := mustCompile(exp)
	g := gre.MustCompile(exp)

	if !reflect.DeepEqual(g.NumSubexp(), r.NumSubexp()) {
		t.Errorf("%#q.NumSubexp() = %v, want %v", exp, r.NumSubexp(), g.NumSubexp())
	}

	if !reflect.DeepEqual(g.SubexpNames(), r.SubexpNames()) {
		t.Errorf("%#q.SubexpNames() = %v, want %v", exp, r.SubexpNames(), g.SubexpNames())
	}

	rm := r.FindStringSubmatchIndex(str)
	gm := g.FindStringSubmatchIndex(str)

	if !reflect.DeepEqual(gm, rm) {
		t.Errorf("%#q.FindSubmatchIndex(%#q) = %v, want %v", exp, str, rm, gm)
	}

	rf := r.FindStringIndex(str)
	gf := g.FindStringIndex(str)

	if !reflect.DeepEqual(gf, rf) {
		t.Errorf("%#q.FindStringIndex(%#q) = %v, want %v", exp, str, rf, gf)
	}

	rs := r.Split(str, -1)
	gs := g.Split(str, -1)
	if !reflect.DeepEqual(rs, gs) {
		t.Errorf("%#q.Split(%#q) = %v, want %v", exp, str, rs, gs)
	}

	r.Longest()
	g.Longest()

	rm = r.FindStringSubmatchIndex(str)
	gm = g.FindStringSubmatchIndex(str)

	if !reflect.DeepEqual(gm, rm) {
		t.Errorf("%#q.FindSubmatchIndex(%#q) [Longest] = %v, want %v", exp, str, rm, gm)
	}
}

func AssertError(t *testing.T, exp string) {
	_, err := compile(exp)
	if err == nil {
		t.Errorf("Compile(%#q) should fail", exp)
	}
}

func AssertEqual(t *testing.T, exp, str string, res []int) {
	r := mustCompile(exp).FindStringSubmatchIndex(str)
	if !reflect.DeepEqual(res, r) {
		t.Errorf("%#q.FindSubmatchIndex(%#q) = %v, want %v", exp, str, r, res)
	}
}

func TestBuiltIn(t *testing.T) {
	file, err := os.Open("./_testdata/builtin.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var exp []string
	var str []string
	for scanner.Scan() {
		q := strings.TrimSpace(scanner.Text())
		if len(q) == 0 {
			continue
		}
		reg := strings.HasPrefix(q, "@")
		if reg {
			q = q[1:]
		}
		s, err := strconv.Unquote(q)
		if err != nil {
			panic(err)
		}
		if reg {
			exp = append(exp, s)
		} else {
			str = append(str, s)
		}
	}

	for _, e := range exp {
		for _, s := range append(str, exp...) {
			AssertBuiltIn(t, e, s)
		}
	}
}

func TestInvalid(t *testing.T) {
	file, err := os.Open("./_testdata/invalid.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		q := strings.TrimSpace(scanner.Text())
		if len(q) == 0 {
			continue
		}
		s, err := strconv.Unquote(q)
		if err != nil {
			panic(err)
		}
		AssertError(t, s)
	}
}

func TestExtended(t *testing.T) {
	file, err := os.Open("./_testdata/extended.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var exp, str string
	for scanner.Scan() {
		q := strings.TrimSpace(scanner.Text())
		if len(q) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(q, "@"):
			s, err := strconv.Unquote(q[1:])
			if err != nil {
				panic(err)
			}
			exp = s
		case q == ">":
			AssertEqual(t, exp, str, nil)
		case strings.HasPrefix(q, ">"):
			var m []int
			for _, n := range strings.Split(q[1:], ",") {
				i, err := strconv.Atoi(strings.TrimSpace(n))
				if err != nil {
					panic(err)
				}
				m = append(m, i)
			}
			AssertEqual(t, exp, str, m)
		default:
			s, err := strconv.Unquote(q)
			if err != nil {
				panic(err)
			}
			str = s
		}
	}
}

func getBenchmarkData() ([]byte, error) {
	file, err := os.Open("./_testdata/アーサー王物語.txt.gz")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(gz)
}

func BenchmarkAny(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := mustCompile(`.*`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkAnyBuiltin(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := MustCompile(`.*`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkLiteral(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := mustCompile(`アーサー`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkLiteralBuiltin(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := MustCompile(`アーサー`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkFirstSpace(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := mustCompile(`^\s*`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkFirstSpaceBuiltin(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := MustCompile(`^\s*`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkAlternation(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := mustCompile(`(?i)arthur|アーサー`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkAlternationBuiltin(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := MustCompile(`(?i)arthur|アーサー`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkHiragana(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := mustCompile(`[\\p{Hiragana}]+`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkHiraganaBuiltin(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := MustCompile(`[\\p{Hiragana}]+`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkURL(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := mustCompile(`([a-zA-Z][a-zA-Z0-9]*)://([^ /]+)(/[^ ]*)?`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}

func BenchmarkURLBuiltin(b *testing.B) {
	data, err := getBenchmarkData()
	if err != nil {
		log.Fatal(err)
	}

	r := MustCompile(`([a-zA-Z][a-zA-Z0-9]*)://([^ /]+)(/[^ ]*)?`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.FindAllSubmatchIndex(data, -1)
	}
}
