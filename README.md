# ![XSX-Logo](doc/xsx-logo.png?raw=true) – eXtended S-eXpressions
[![Build Status](https://travis-ci.org/fractalqb/xsx.svg)](https://travis-ci.org/fractalqb/xsx)

`import "git.fractalqb.de/fractalqb/qblog"`

---

Package XSX provides tools for parsing something I call eXtended
S-eXpressions.  Extended means the following things compared to
[SEXP S-expressions](https://people.csail.mit.edu/rivest/sexp.html):

1. Nested structures are delimited by balanced braces '()', '[]' or
   '{}’ – not only by '()'.

2. XSX provides a notation for "Meta Values", i.e. XSXs that provide
   some sort of meta information that is not part of the "normal"
   data.

On the other hand some properties from SEXP were dropped, e.g. typing
of the so called "octet strings". Things like that are completely left
to the application.

## Somewhat more formal description

Frist of all, XSX is not about datatypes, in this it is comparable to
e.g. XML (No! don't leave… its much simpler). Instead its building
block of is the _atom_, i.e. nothing else than a sequence of
characters, aka a 'string'. Atoms come as _quoted atoms_ and as
_unquoted atoms_. One needs to quote an atom when the atom's string
contains characters that have a special meaning in XSX: ()[]{}\ and
white-space.

### Regexp style definition of Atom

    atom     := nq-atom | q-atom
    nq-atom  := ([^()[]{}]|\s)+
	q-atom   := "([^"\]|(\")(\\))+"
	XSX      := atom

I.e. `x` is an atom and `foo`, `bar` and `baz` are atoms. An atom that
contains a '"' or '\' would be `"quote: \" and backslash: \\ in a
quoted atom"`.  Also `"("` is an atom but `(` is not an atom. We need
'(' for other things!

### Sequences now BNF Style

Each atom is an XSX. From XSX one can build sequences:

    xsxs ::= XSX | XSX xsxs
    seq1 ::= '(' ')' | '(' xsxs ')'
    seq2 ::= '[' ']' | '[' xsxs ']'
    seq3 ::= '{' '}' | '{' xsxs '}'
    XSX  ::= atom | seq1 | seq2 | seq3

### Out-Of-Band Information with Meta XSXs

You can prefix each XSX with a backslash '\' to make that expression a
meta-expression. A meta-expression is not considered to be a XSX,
i.e. you cannot create meta-meta-expressions or
meta-meta-meta-expressions… hmm… and not event
meta-meta-meta-meta-expressions! I think it became clear?

E.g. `\4711` is a meta-atom and `\{foo 1 bar false baz 3.1415}` is a
meta-sequence. What _meta_ means is completely up to the
application. Imagine e.g. `(div hiho)` and `(div \{class green} hiho)`
to be a translation from `<div>hiho</div>` and `<div
class="green">hiho</div>`.

## Rationale

None! … despite the fact that I found it to be fun – and useful in
some situations.

Because its so simple it is easy to use the `PullParser` as a
tokenizer to build customized parsers for proprietary data
files. E.g. see the `table` sub-package.
