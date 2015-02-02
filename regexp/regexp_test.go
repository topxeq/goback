package regexp

import (
	"bufio"
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
	file, err := os.Open("./builtin.txt")
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
	file, err := os.Open("./invalid.txt")
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
	file, err := os.Open("./extended.txt")
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
