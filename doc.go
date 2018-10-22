// Package xsx provides tools for parsing so called eXtended S-eXpressions.
// Extended means the following things compared to
// https://people.csail.mit.edu/rivest/sexp.html:
//
// Nested structures are delimited by balanced braces '()', '[]' or '{}’ – not
// only by '()'.
//
// XSX provides a notation for "Meta Values", i.e. XSX that provide some sort
// of meta information that is not part of the "normal" data.
//
// On the other hand some properties from SEXP were dropped, e.g. typing
// of the so called "octet strings". Things like that are completely left
// to the application.
package xsx

//go:generate versioner -pkg xsx ./VERSION ./version.go
