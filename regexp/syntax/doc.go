/*
Package syntax parses regular expressions into parse trees and compiles
parse trees into programs.

Syntax

This package supports following syntax in addition to the golang built-in regexp.


Grouping:
  (?>re)         atomic group; non-capturing
  (?=re)         lookahead; non-capturing
  (?!re)         negative lookahead; non-capturing
  (?#comment)    comment

Repetitions:
  x*+            zero or more x, possessive
  x++            one or more x, possessive
  x?+            zero or one x, possessive
  x{n,m}+        n or n+1 or ... or m x, possessive
  x{n,}+         n or more x, possessive
  x{n}+          exactly n x, possessive

Back reference:
  \kN            refer to numbered capturing
  \kName         refer to named capturing
  \k{N}          refer to numbered capturing
  \k{Name}       refer to named capturing

*/
package syntax
