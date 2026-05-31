//
// The YZ Interpreter
// Licensed under the GPL v3.0
//
//

package main

// Imports

import (
	"fmt"
	"log"
	"os"
	"unicode"
)

// AST Nodes

type Num struct {
	Value int
}

type Var struct {
	Name string
}

type Add struct {
	Left  any
	Right any
}

type Sub struct {
	Left  any
	Right any
}

type Mul struct {
	Left  any
	Right any
}

type Let struct {
	Name  string
	Value any
}

type Print struct {
	Expr any
}

type Program struct {
	statements []any
}

// Tokenizer

type Token struct {
	Type  string
	Value string
}

func lexer(src string) []Token {
	var i int = 0
	var tokens []Token = []Token{}

	for i < len(src) {
		var c = src[i]
		if unicode.IsSpace(rune(c)) {
			i += 1
		} else if unicode.IsDigit(rune(c)) {
			var j = i
			for j < len(src) && unicode.IsDigit(rune(src[j])) {
				j += 1
			}
			tokens = append(tokens, Token{"NUMBER", src[i:j]})
			i = j
		} else if unicode.IsLetter(rune(c)) || c == '_' {
			var j = i
			for j < len(src) && (unicode.IsLetter(rune(src[j])) || unicode.IsDigit(rune(src[j])) || src[j] == '_') {
				j += 1
			}
			var word = src[i:j]
			switch word {
			case "let":
				tokens = append(tokens, Token{"LET", word})
			case "println":
				tokens = append(tokens, Token{"PRINTLN", word})
			default:
				tokens = append(tokens, Token{"IDENT", word})
			}
			i = j
		} else if c == '+' {
			tokens = append(tokens, Token{"PLUS", string(c)})
			i += 1
		} else if c == '-' {
			tokens = append(tokens, Token{"MINUS", string(c)})
			i += 1
		} else if c == '*' {
			tokens = append(tokens, Token{"MUL", string(c)})
			i += 1
		} else if c == '=' {
			tokens = append(tokens, Token{"EQUAL", string(c)})
			i += 1
		} else if c == ';' {
			tokens = append(tokens, Token{"SEMI", string(c)})
			i += 1
		} else if c == '(' {
			tokens = append(tokens, Token{"LPAREN", string(c)})
			i += 1
		} else if c == ')' {
			tokens = append(tokens, Token{"RPAREN", string(c)})
			i += 1
		} else {
			log.Fatalf("SyntaxError: Unexpected character: %s", string(c))
			os.Exit(0)
		}
	}

	tokens = append(tokens, Token{"EOF", ""})

	return tokens
}

func main() {
	data, err := os.ReadFile("examples/1.yz")
	if err != nil {
		log.Fatal(err)
	}

	content := string(data)

	fmt.Println(lexer(content))
}
