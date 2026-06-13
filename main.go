// SPDX-FileCopyrightText: 2026 ark
// SPDX-License-Identifier: GPL-3.0-or-later
//
// The YZ Interpreter
// Licensed under the GNU General Public License v3.0
//

package main

// Imports

import (
	"bufio"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"

	. "modernc.org/tk9.0"
	_ "modernc.org/tk9.0/themes/azure"
)

// AST Nodes

type Num struct {
	Value int
}

type Str struct {
	Value string
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
	Name       string
	Value      any
	Visibility string
}

type Print struct {
	Expr any
}

type IfStmt struct {
	Condition any
	Then      []any
	Else      []any
}

type WhileLoop struct {
	Condition any
	Contents  []any
}

type GoThruLoop struct {
	ArrayVar string
	Contents []any
	IterVar  string
}

type Function struct {
	Name       string
	Parameters map[string]any
	Contents   []any
	Visibility string
}

type YZVariable struct {
	Value      any
	Visibility string
}

type FuncCallStatement struct {
	Name       string
	Parameters map[string]any
}

type FuncCallExpr struct {
	Name       string
	Parameters map[string]any
}

type Program struct {
	statements []any
	variables  map[string]YZVariable
	functions  map[string]Function
}

type Return struct {
	Value any
}

type Comparison struct {
	Left     any
	Operator string
	Right    any
}

type ImportStmt struct {
	library string
}

type YZInvokeStmt struct {
	func_to_invoke string
	return_var     string
	Parameters     map[string]any
}

type BreakStmt struct{}

// Custom error types

var ErrorBreak = errors.New("break")

// Tokenizer

type Token struct {
	Type  string
	Value string
}

var keywords = map[string]string{
	"let": "LET",
	"println": "PRINTLN",
	"if": "IF",
	"func": "FUNC",
	"return": "RETURN",
	"else": "ELSE",
	"import": "IMPORT",
	"public": "PUBLIC",
	"private": "PRIVATE",
	"_yz_invoke": "YZ_INVOKE",
	"while": "WHILE",
	"break": "BREAK",
	"gothru": "GOTHRU",
	"as": "AS",
	"or": "OR",
}

func atoi(s string) (int, error) {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c - '0')
	}
	return n, nil
}

func lexer(src string, verbose *bool) []Token {
	var i int = 0
	var tokens []Token = make([]Token, 0, len(src)/2)

	for i < len(src) {
		var c = src[i]

		if unicode.IsSpace(rune(c)) {
			i += 1
		} else if unicode.IsDigit(rune(c)) || (c == '-' && unicode.IsDigit(rune(src[i+1]))) {
			var j = i

			if c == '-' && unicode.IsDigit(rune(src[i+1])) {
				j += 1
			}

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
			if w, ok := keywords[word]; ok {
				tokens = append(tokens, Token{w, word})
			} else {
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
		} else if c == '=' && src[i+1] == '=' {
			tokens = append(tokens, Token{"DOUBLE_EQUAL", string(src[i : i+1])})
			i += 2
		} else if c == '>' {
			if src[i+1] == '=' {
				tokens = append(tokens, Token{"GREATER_EQUAL", string(src[i : i+1])})
				i += 2
			} else {
				tokens = append(tokens, Token{"GREATER", string(c)})
				i += 1
			}
		} else if c == '<' {
			if src[i+1] == '=' {
				tokens = append(tokens, Token{"LESS_THAN_EQUAL", string(src[i : i+1])})
				i += 2
			} else {
				tokens = append(tokens, Token{"LESS_THAN", string(c)})
				i += 1
			}
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
		} else if c == '"' {
			tokens = append(tokens, Token{"QUOTE", string(c)})
			i += 1
			var j = i
			for j < len(src) && src[j] != '"' {
				j += 1
			}
			tokens = append(tokens, Token{"STR", src[i:j]})
			tokens = append(tokens, Token{"QUOTE", "\""})
			i = j + 1
		} else if c == ',' {
			tokens = append(tokens, Token{"COMMA", string(c)})
			i += 1
		} else if c == '{' {
			tokens = append(tokens, Token{"LBRACE", string(c)})
			i += 1
		} else if c == '}' {
			tokens = append(tokens, Token{"RBRACE", string(c)})
			i += 1
		} else if c == '#' {
			tokens = append(tokens, Token{"HASH", string(c)})
			i += 1
		} else if c == '/' && i+1 < len(src) && src[i+1] == '/' {
			for i < len(src) && src[i] != '\n' {
				i += 1
			}
		} else {
			log.Fatalf("SyntaxError: Unexpected character: %s", string(c))
		}
	}

	tokens = append(tokens, Token{"EOF", ""})

	if *verbose {
		fmt.Println("Lexer Tokens:", tokens)
	}

	return tokens
}

// Parser

type Parser struct {
	tokens    []Token
	pos       int
	variables map[string]YZVariable
	functions map[string]Function
}

func new_parser(tokens []Token, variables map[string]YZVariable, functions map[string]Function) *Parser {
	return &Parser{tokens: tokens, pos: 0, variables: variables, functions: functions}
}

func (p *Parser) cur() Token {
	return p.tokens[p.pos]
}

func (p *Parser) peek_next() Token {
	return p.tokens[p.pos+1]
}

func (p *Parser) eat(typ string) Token {
	if p.cur().Type != typ {
		log.Fatalf("Expected %s, got %s %s.", typ, p.cur().Type, p.cur().Value)
	}
	tok := p.cur()
	p.pos += 1
	return tok
}

func (p *Parser) parse() Program {
	stmts := []any{}
	for p.cur().Type != "EOF" {
		stmts = append(stmts, p.statement())
	}
	return Program{stmts, p.variables, p.functions}
}

func (p *Parser) statement() any {
	if p.cur().Type == "LET" {
		return p.let_statement()
	} else if p.cur().Type == "PRINTLN" {
		return p.println_statement()
	} else if p.cur().Type == "IF" {
		return p.if_statement()
	} else if p.cur().Type == "FUNC" {
		return p.func_statement()
	} else if p.cur().Type == "RETURN" {
		return p.return_statement()
	} else if p.cur().Type == "IMPORT" {
		return p.import_statement()
	} else if p.cur().Type == "YZ_INVOKE" {
		return p.yz_invoke_statement()
	} else if p.cur().Type == "IDENT" {
		if p.peek_next().Type == "LPAREN" {
			return p.func_call_statement()
		} else {
			return p.expr()
		}
	} else if p.cur().Type == "WHILE" {
		return p.while_statement()
	} else if p.cur().Type == "BREAK" {
		return p.break_statement()
	} else if p.cur().Type == "GOTHRU" {
		return p.gothru_statement()
	} else {
		log.Fatalf("Unexpected statement token: %s", p.cur().Type)
		os.Exit(0)
		return 0
	}
}

func (p *Parser) let_statement() Let {
	p.eat("LET")
	name := p.eat("IDENT").Value

	visibility := "private"
	if p.cur().Type == "PUBLIC" {
		visibility = "public"
		p.eat(p.cur().Type)
	} else if p.cur().Type == "PRIVATE" {
		visibility = "private"
		p.eat(p.cur().Type)
	}

	p.eat("EQUAL")
	value := p.expr()
	p.eat("SEMI")
	return Let{name, value, visibility}
}

func (p *Parser) println_statement() Print {
	p.eat("PRINTLN")
	p.eat("LPAREN")
	expr := p.expr()
	p.eat("RPAREN")
	p.eat("SEMI")
	return Print{expr}
}

func (p *Parser) if_statement() IfStmt {
	expressions := []Comparison{}

	p.eat("IF")
	expr1 := p.expr()
	comparison_operator := Token{}
	switch p.cur().Type {
	case "DOUBLE_EQUAL", "LESS_THAN_EQUAL", "LESS_THAN", "GREATER_EQUAL", "GREATER":
		comparison_operator = p.eat(p.cur().Type)
	default:
		log.Fatalf("%s is not a comparison operator", p.cur().Type)
	}
	expr2 := p.expr()

	if p.cur().Type == "OR" {
		p.eat("OR")
		expressions = append(expressions, Comparison{Left: expr1, Operator: comparison_operator.Type, Right: expr2})

		for p.cur().Type != "LBRACE" {
			expr1 := p.expr()
			comparison_operator := Token{}
			switch p.cur().Type {
			case "DOUBLE_EQUAL", "LESS_THAN_EQUAL", "LESS_THAN", "GREATER_EQUAL", "GREATER":
				comparison_operator = p.eat(p.cur().Type)
			default:
				log.Fatalf("%s is not a comparison operator", p.cur().Type)
			}
			expr2 := p.expr()
			expressions = append(expressions, Comparison{Left: expr1, Operator: comparison_operator.Type, Right: expr2})
			if p.cur().Type == "OR" {
				p.eat("OR")
			}
		}
	}

	p.eat("LBRACE")

	thenStmts := []any{}
	for p.cur().Type != "RBRACE" {
		thenStmts = append(thenStmts, p.statement())
	}
	p.eat("RBRACE")

	elseStmts := []any{}
	if p.cur().Type == "ELSE" {
		p.eat("ELSE")
		p.eat("LBRACE")

		for p.cur().Type != "RBRACE" {
			elseStmts = append(elseStmts, p.statement())
		}

		p.eat("RBRACE")
	}

	if len(expressions) > 0 {
		return IfStmt{expressions, thenStmts, elseStmts}
	} else {
		return IfStmt{Comparison{Left: expr1, Operator: comparison_operator.Type, Right: expr2}, thenStmts, elseStmts}
	}
}

func (p *Parser) while_statement() WhileLoop {
	p.eat("WHILE")
	expr1 := p.expr()
	comparison_operator := Token{}
	switch p.cur().Type {
	case "DOUBLE_EQUAL", "LESS_THAN_EQUAL", "LESS_THAN", "GREATER_EQUAL", "GREATER":
		comparison_operator = p.eat(p.cur().Type)
	default:
		log.Fatalf("%s is not a comparison operator", p.cur().Type)
	}
	expr2 := p.expr()

	p.eat("LBRACE")

	stmts := []any{}
	for p.cur().Type != "RBRACE" {
		stmts = append(stmts, p.statement())
	}
	p.eat("RBRACE")

	elseStmts := []any{}
	if p.cur().Type == "ELSE" {
		p.eat("ELSE")
		p.eat("LBRACE")

		for p.cur().Type != "RBRACE" {
			elseStmts = append(elseStmts, p.statement())
		}

		p.eat("RBRACE")
	}

	return WhileLoop{Comparison{Left: expr1, Operator: comparison_operator.Type, Right: expr2}, stmts}
}

func (p *Parser) gothru_statement() GoThruLoop {
	p.eat("GOTHRU")
	array := p.eat("IDENT").Value
	p.eat("AS")
	itervar := p.eat("IDENT").Value

	p.eat("LBRACE")

	stmts := []any{}
	for p.cur().Type != "RBRACE" {
		stmts = append(stmts, p.statement())
	}
	p.eat("RBRACE")

	return GoThruLoop{array, stmts, itervar}
}

func (p *Parser) func_statement() Function {
	args := map[string]any{}

	p.eat("FUNC")
	name := p.eat("IDENT").Value
	p.eat("LPAREN")
	for p.cur().Type != "RPAREN" {
		if p.cur().Type == "HASH" {

			if p.peek_next().Type == "IDENT" {
				p.eat("HASH")
				if p.eat("IDENT").Value == "arbitrary_params_allowed" {
					args["_yz_arbitrary_params_allowed_"] = "YES"
				} else {
					log.Fatalf("Unknown hash parameter on function %s.", name)
				}
				if p.cur().Type != "RPAREN" {
					p.eat("COMMA")
				}
			}

		} else {

			args[p.eat("IDENT").Value] = "null"
			if p.cur().Type != "RPAREN" {
				p.eat("COMMA")
			}

		}
	}
	p.eat("RPAREN")

	visibility := "private"
	if p.cur().Type == "PUBLIC" {
		visibility = "public"
		p.eat(p.cur().Type)
	} else if p.cur().Type == "PRIVATE" {
		visibility = "private"
		p.eat(p.cur().Type)
	}

	p.eat("LBRACE")
	funcStmts := []any{}
	for p.cur().Type != "RBRACE" {
		funcStmts = append(funcStmts, p.statement())
	}
	p.eat("RBRACE")
	return Function{name, args, funcStmts, visibility}
}

func (p *Parser) func_call_statement() FuncCallStatement {
	args := map[string]any{}
	func_name := p.eat("IDENT").Value
	p.eat("LPAREN")
	for p.cur().Type != "RPAREN" {
		param_name := p.eat("IDENT").Value
		args[param_name] = p.expr()
		if p.cur().Type != "RPAREN" {
			p.eat("COMMA")
		}
	}
	p.eat("RPAREN")
	p.eat("SEMI")
	return FuncCallStatement{func_name, args}
}

func (p *Parser) func_call_expr() FuncCallExpr {
	args := map[string]any{}
	func_name := p.eat("IDENT").Value
	p.eat("LPAREN")
	for p.cur().Type != "RPAREN" {
		param_name := p.eat("IDENT").Value
		args[param_name] = p.expr()
		if p.cur().Type != "RPAREN" {
			p.eat("COMMA")
		}
	}
	p.eat("RPAREN")
	return FuncCallExpr{func_name, args}
}

func (p *Parser) return_statement() Return {
	p.eat("RETURN")
	value := p.expr()
	p.eat("SEMI")
	return Return{value}
}

func (p *Parser) import_statement() ImportStmt {
	p.eat("IMPORT")
	library := p.eat("IDENT").Value
	p.eat("SEMI")
	return ImportStmt{library}
}

func (p *Parser) yz_invoke_statement() YZInvokeStmt {
	p.eat("YZ_INVOKE")
	p.eat("LPAREN")
	func_to_invoke := p.eat("IDENT").Value
	p.eat("COMMA")
	return_var := p.eat("IDENT").Value
	params := make(map[string]any)
	if p.cur().Type != "RPAREN" {
		p.eat("COMMA")
		for p.cur().Type != "RPAREN" {
			param_name := p.eat("IDENT").Value
			param_value := p.expr()
			params[param_name] = param_value
			if p.cur().Type != "RPAREN" {
				p.eat("COMMA")
			}
		}
	}
	p.eat("RPAREN")
	p.eat("SEMI")
	return YZInvokeStmt{func_to_invoke, return_var, params}
}

func (p *Parser) break_statement() BreakStmt {
	p.eat("BREAK")
	p.eat("SEMI")
	return BreakStmt{}
}

func (p *Parser) expr() any {
	if p.cur().Type == "QUOTE" {
		p.eat("QUOTE")
		str := p.eat("STR").Value
		p.eat("QUOTE")
		return Str{str}
	} else if p.cur().Type == "IDENT" && p.peek_next().Type == "LPAREN" {
		return p.func_call_expr()
	}

	return p.add_expr()
}

func (p *Parser) add_expr() any {
	node := p.mul_expr()
	for p.cur().Type == "PLUS" || p.cur().Type == "MINUS" {
		operator := p.cur().Type
		p.eat(operator)
		right := p.mul_expr()
		if operator == "PLUS" {
			node = Add{node, right}
		} else {
			node = Sub{node, right}
		}
	}
	return node
}

func (p *Parser) mul_expr() any {
	node := p.primary()
	for p.cur().Type == "MUL" {
		p.eat("MUL")
		right := p.primary()
		node = Mul{node, right}
	}
	return node
}

func func_call_and_return(call FuncCallExpr, variables map[string]YZVariable, functions map[string]Function) any {
	if fn, exists := functions[call.Name]; exists {
		func_vars := make(map[string]YZVariable)
		for key, value := range variables {
			func_vars[key] = value
		}
		for param, param_value := range call.Parameters {
			if _, ok := fn.Parameters[param]; ok {
				func_vars[param] = YZVariable{eval_expr(param_value, variables, functions), "private"}
			} else {
				if fn.Parameters["_yz_arbitrary_params_allowed_"] == "YES" {
					func_vars[param] = YZVariable{eval_expr(param_value, variables, functions), "private"}
				} else {
					log.Fatalf("Call to function %s failed due to non-existent parameter %s without _yz_arbitrary_params_allowed_ flag.", call.Name, param)
				}
			}
		}
		for name := range fn.Parameters {
			if name != "_yz_arbitrary_params_allowed_" {
				if _, ok := call.Parameters[name]; !ok {
					log.Fatalf("Missing required parameter %s on function %s", name, fn.Name)
				}
			}
		}
		for _, stmt := range fn.Contents {
			if ret, ok := stmt.(Return); ok {
				return eval_expr(ret.Value, func_vars, functions)
			}
			run_statement(stmt, func_vars, functions)
		}
		return ""
	} else {
		log.Fatalf("Call to non-existent function %s.", call.Name)
		return ""
	}
}

func (p *Parser) primary() any {
	tok := p.cur()

	switch tok.Type {
	case "NUMBER":
		p.eat("NUMBER")
		val, _ := atoi(tok.Value)
		return Num{val}
	case "IDENT":
		p.eat("IDENT")
		return Var{tok.Value}
	case "LPAREN":
		p.eat("LPAREN")
		node := p.expr()
		p.eat("RPAREN")
		return node
	case "QUOTE":
		value := ""
		p.eat("QUOTE")
		if p.cur().Type == "STR" {
			value = p.cur().Value
		}
		p.eat("STR")
		p.eat("QUOTE")
		return Str{value}
	default:
		log.Fatalf("Unexpected token in primary (%s, %s)", tok.Type, tok.Value)
		return 0
	}
}

func run(program *Program) {
	for _, stmt := range program.statements {
		if _, err := run_statement(stmt, program.variables, program.functions); err == ErrorBreak {
			break
		}
	}
}

func run_statement(stmt any, variables map[string]YZVariable, functions map[string]Function) (any, error) {
	switch s := stmt.(type) {
	case Let:
		value := eval_expr(s.Value, variables, functions)
		variables[s.Name] = YZVariable{value, s.Visibility}
		return "", nil
	case Print:
		value := eval_expr(s.Expr, variables, functions)
		fmt.Println(value)
		return "", nil
	case IfStmt:
		condition := eval_expr(s.Condition, variables, functions) == "true"
		if condition {
			for _, thenStmt := range s.Then {
				if _, err := run_statement(thenStmt, variables, functions); err != nil {
					return "", err
				}
			}
		} else {
			for _, elseStmt := range s.Else {
				if _, err := run_statement(elseStmt, variables, functions); err != nil {
					return "", err
				}
			}
		}
		return "", nil
	case WhileLoop:
		for eval_expr(s.Condition, variables, functions) == "true" {
			for _, thenStmt := range s.Contents {
				if _, err := run_statement(thenStmt, variables, functions); err == ErrorBreak {
					return "", nil // break outta the while loop
				}
			}
		}
		return "", nil
	case Function:
		functions[s.Name] = s
		return "", nil
	case FuncCallStatement:
		if fn, exists := functions[s.Name]; exists {
			func_vars := make(map[string]YZVariable)

			for key, value := range variables {
				func_vars[key] = value
			}

			for param, param_value := range s.Parameters {
				if _, ok := fn.Parameters[param]; ok {
					func_vars[param] = YZVariable{eval_expr(param_value, variables, functions), "private"}
				} else {
					if functions[s.Name].Parameters["_yz_arbitrary_params_allowed_"] == "YES" {
						func_vars[param] = YZVariable{eval_expr(param_value, variables, functions), "private"}
					} else {
						log.Fatalf("Call to function %s failed due to non-existent parameter %s without _yz_arbitrary_params_allowed_ flag.", param, s.Name)
					}
				}
			}

			for name := range functions[s.Name].Parameters {
				if name != "_yz_arbitrary_params_allowed_" {
					if _, ok := s.Parameters[name]; !ok {
						log.Fatalf("Missing required parameter %s", name)
					}
				}
			}

			for _, stmt := range fn.Contents {
				if _, err := run_statement(stmt, func_vars, functions); err != nil {
					return "", err
				}
			}
		} else {
			log.Fatalf("Call to non-existent function %s.", s.Name)
		}
		return "", nil
	case Return:
		return eval_expr(s.Value, variables, functions), nil
	case ImportStmt:
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		lib_path := dir + ".yz/libs/" + s.library + ".yz"
		var library_contents []byte

		if _, err := os.Stat(lib_path); err == nil {
			library_contents, err = os.ReadFile(lib_path)
			if err != nil {
				log.Fatal(err)
			}
		} else if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("Trying to import non-existent library %s", s.library)
		} else {
			log.Fatal(err)
		}

		verbose := false
		tokens := lexer(string(library_contents), &verbose)
		parser := new_parser(tokens, make(map[string]YZVariable), make(map[string]Function))
		program := parser.parse()
		run(&program)
		for _, f := range program.functions {
			if f.Visibility == "public" {
				functions[f.Name] = Function{f.Name, f.Parameters, f.Contents, f.Visibility}
			}
		}
		for k, v := range program.variables {
			if v.Visibility == "public" {
				variables[k] = YZVariable{v.Value, "public"}
			}
		}
		return "", nil
	case YZInvokeStmt:
		parameters := make(map[string]any)
		for k, v := range s.Parameters {
			parameters[k] = eval_expr(v, variables, functions)
		}
		variables[s.return_var] = YZVariable{handle_yz_invoke(s, parameters, variables, functions), "private"}
		return "", nil
	case BreakStmt:
		return "", ErrorBreak
	case GoThruLoop:
		array := variables[s.ArrayVar]
		switch a := array.Value.(type) {
		case []string:
			for _, i := range a {
				variables[s.IterVar] = YZVariable{i, "private"}
				for _, thenStmt := range s.Contents {
					if _, err := run_statement(thenStmt, variables, functions); err == ErrorBreak {
						return "", nil
					}
				}
				variables[s.IterVar] = YZVariable{"", "private"}
			}
		case []any:
			for _, i := range a {
				variables[s.IterVar] = YZVariable{i, "private"}
				for _, thenStmt := range s.Contents {
					if _, err := run_statement(thenStmt, variables, functions); err == ErrorBreak {
						return "", nil
					}
				}
				variables[s.IterVar] = YZVariable{"", "private"}
			}
		default:
			log.Fatalf("Unexpected array type in gothru statement: %s", reflect.TypeOf(array))
		}

		return "", nil
	default:
		log.Fatalf("Unknown statement:\nType: %s\nValue: %s\n\n", reflect.TypeOf(s).String(), s)
		return "", nil
	}
}

func handle_yz_invoke(s YZInvokeStmt, params map[string]any, variables map[string]YZVariable, functions map[string]Function) any {
	function := s.func_to_invoke

	if strings.HasPrefix(function, "_yz_cmd_") {
		function = strings.Replace(function, "_yz_cmd_", "", 1)
	}

	switch function {
	case "rand_num":
		min, _ := atoi(params["min"].(string))
		max, _ := atoi(params["max"].(string))
		return strconv.Itoa(min + rand.IntN(max-min+1))
	case "guitk_activate_theme":
		ActivateTheme(params["theme"].(string))
		return ""
	case "guitk_pack":
		paramss := make(map[string]string)

		for _, v := range strings.Split(params["widget"].(string), "|;|") {
			parts := strings.Split(v, "=======")
			if len(parts) == 2 {
				paramss[parts[0]] = parts[1]
			}
		}

		var widget any = nil

		switch paramss["widget"] {
		case "label":
			widget = Label(Txt(paramss["text"]))
			Pack(widget.(*LabelWidget))
		case "inputbox":
			width, _ := atoi(paramss["width"])
			Pack(TEntry(Textvariable(""), Background(White), Width(width)))
		case "textbox":
			widget = Text(Height(14), Width(60))
			Pack(widget.(*TextWidget))
		case "button":
			Pack(TButton(Txt(paramss["text"]), Command(func() {
				params := make(map[string]any)
				var bracket_rgx = regexp.MustCompile(`\((.*?)\)`)
				if len(bracket_rgx.FindStringSubmatch(paramss["text"])) < 1 {
					params["btn_text"] = Str{paramss["text"]}
				} else {
					params["btn_text"] = Str{bracket_rgx.FindStringSubmatch(paramss["text"])[1]}
				}
				run_statement(FuncCallStatement{paramss["onClickFunc"], params}, variables, functions)
			})))
		}

		return widget
	case "guitk_open_file":
		path := GetOpenFile()[0]
		return path
	case "guitk_read_file":
		file, err := os.Open(params["path"].(string))
		if err != nil {
			log.Fatalf("Error opening file: %s", err)
			return "ERR_OPEN_FILE"
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("Error reading file: %s", err)
			return "ERR_READ_FILE"
		}

		return string(data)
	case "guitk_insert":
		params["textbox"].(*TextWidget).Insert("end", params["text"].(string))
		return 1
	case "guitk_clear":
		params["textbox"].(*TextWidget).Delete("1.0", "end")
		return 1
	case "guitk_set_content":
		params["label"].(*LabelWidget).Configure(Txt(params["text"].(string)))
		return 1
	case "guitk_get_content":
		switch w := params["widget"].(type) {
		case *TextWidget:
			return w.Get("1.0", "end-1c")
		case *LabelWidget:
			return w.Txt()
		default:
			return ""
		}
	case "fs_write_file":
		f, err := os.Create(params["path"].(string))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.WriteString(params["content"].([]string)[0])
		return 1
	case "guitk_loop":
		App.Wait()
		return ""
	case "guitk_set_title":
		App.WmTitle(params["title"].(string))
		return ""
	case "guitk_quit_gui":
		Destroy(App)
		return 200
	case "startswith":
		return strconv.FormatBool(strings.HasPrefix(params["s"].(string), params["prefix"].(string)))
	case "http_request":
		req, err := http.NewRequest("GET", params["url"].(string), nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		return string(body)
	case "parse_html":
		htmlcontent := params["html"].(string)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlcontent))
		if err != nil {
			log.Fatal(err)
		}

		result := []string{}

		doc.Find("*").Each(func(i int, s *goquery.Selection) {
			result = append(result, goquery.NodeName(s))
			result = append(result, strings.TrimSpace(s.Text()))
			for _, n := range s.Nodes {
				for _, a := range n.Attr {
					result = append(result, a.Key)
					result = append(result, a.Val)
				}
			}
			result = append(result, ".")
		})

		return result
	case "make_list":
		return []any{}
	case "make_dictionary":
		return make(map[string]any)
	case "dict_set":
		switch dict := params["dict"].(type) {
		case map[string]any:
			dict[params["key"].(string)] = params["value"]
			return dict
		default:
			log.Fatalf("Failed: Trying to set dictionary value on dictionary of type %T", params["dict"])
		}
		return ""
	case "dict_get":
		switch dict := params["dict"].(type) {
		case map[string]any:
			return dict[params["key"].(string)]
		default:
			log.Fatalf("Failed: Trying to get value from dictionary of type %T", params["dict"])
		}
		return ""
	case "append_to_list":
		switch lst := params["list"].(type) {
		case []any:
			return append(lst, params["value"].(string))
		default:
			log.Fatalf("Failed: Trying to append value to array of type %T.", params["list"])
		}
		return ""
	case "valuefromindex":
		index, err := atoi(params["index"].(string))
		if err != nil {
			panic(err)
		}

		switch lst := params["list"].(type) {
		case []any:
			if index < len(lst) {
				return lst[index]
			} else {
				return ""
			}
		default:
			log.Fatalf("Failed: Trying to get value from array of type %T.", params["list"])
		}
		return ""
	case "listlength":
		switch lst := params["list"].(type) {
		case []any:
			return strconv.Itoa(len(lst))
		default:
			log.Fatalf("Failed: Trying to get length of array of type %T.", params["list"])
		}
		return 0
	case "get_input":
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		return scanner.Text()
	case "null":
		return nil
	case "runshellcmd":
		cmd := exec.Command("bash", "-c", params["cmd"].(string))
		out, err := cmd.Output()
		if err != nil {
			log.Fatalln("Error:", err)
		}
		return string(out)
	default:
		return ""
	}
}

func eval_random_expr(s string, variables map[string]YZVariable, functions map[string]Function) any {
	verbose := false
	tokens := lexer(s, &verbose)
	parser := new_parser(tokens, variables, functions)
	expr := parser.expr()
	return eval_expr(expr, variables, functions)
}

func eval_expr(expr any, variables map[string]YZVariable, functions map[string]Function) any {
	switch e := expr.(type) {
	case Num:
		return strconv.Itoa(e.Value)
	case Str:
		re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

		string_value := re.ReplaceAllStringFunc(e.Value, func(match string) string {
			submatches := re.FindStringSubmatch(match)
			if len(submatches) > 1 {
				return eval_random_expr(submatches[1], variables, functions).(string) // second value of submatches is the content inside of the braces
			}
			return match
		})

		return string_value
	case Var:
		if _, ok := variables[e.Name]; ok {
			return variables[e.Name].Value
		} else {
			log.Fatalf("Reference of non-existent variable %s in expression %s", e.Name, expr)
		}
	case Comparison:
		expr1 := eval_expr(e.Left, variables, functions)
		expr2 := eval_expr(e.Right, variables, functions)

		condition := false
		switch e.Operator {
		case "DOUBLE_EQUAL":
			condition = expr1 == expr2
		case "LESS_THAN_EQUAL":
			if e1, err := atoi(expr1.(string)); err == nil {
				if e2, err := atoi(expr2.(string)); err == nil {
					condition = e1 <= e2
				} else {
					log.Fatalf("2nd expression in operation not a number.")
				}
			} else {
				log.Fatalf("1st expression in operation not a number.")
			}
		case "GREATER_EQUAL":
			if e1, err := atoi(expr1.(string)); err == nil {
				if e2, err := atoi(expr2.(string)); err == nil {
					condition = e1 >= e2
				} else {
					log.Fatalf("2nd expression in operation not a number.")
				}
			} else {
				log.Fatalf("1st expression in operation not a number.")
			}
		case "LESS_THAN":
			if e1, err := atoi(expr1.(string)); err == nil {
				if e2, err := atoi(expr2.(string)); err == nil {
					condition = e1 < e2
				} else {
					log.Fatalf("2nd expression in operation not a number.")
				}
			} else {
				log.Fatalf("1st expression in operation not a number.")
			}
		case "GREATER":
			if e1, err := atoi(expr1.(string)); err == nil {
				if e2, err := atoi(expr2.(string)); err == nil {
					condition = e1 > e2
				} else {
					log.Fatalf("2nd expression in operation not a number.")
				}
			} else {
				log.Fatalf("1st expression in operation not a number.")
			}
		default:
			log.Fatalf("%s is not a comparison operator", e.Operator)
		}

		return strconv.FormatBool(condition)
	case []Comparison:
		conditions := []bool{}

		for _, c := range e {
			expr1 := eval_expr(c.Left, variables, functions)
			expr2 := eval_expr(c.Right, variables, functions)

			switch c.Operator {
			case "DOUBLE_EQUAL":
				conditions = append(conditions, expr1 == expr2)
			case "LESS_THAN_EQUAL":
				if e1, err := atoi(expr1.(string)); err == nil {
					if e2, err := atoi(expr2.(string)); err == nil {
						conditions = append(conditions, e1 <= e2)
					} else {
						log.Fatalf("2nd expression in operation not a number.")
					}
				} else {
					log.Fatalf("1st expression in operation not a number.")
				}
			case "GREATER_EQUAL":
				if e1, err := atoi(expr1.(string)); err == nil {
					if e2, err := atoi(expr2.(string)); err == nil {
						conditions = append(conditions, e1 >= e2)
					} else {
						log.Fatalf("2nd expression in operation not a number.")
					}
				} else {
					log.Fatalf("1st expression in operation not a number.")
				}
			case "LESS_THAN":
				if e1, err := atoi(expr1.(string)); err == nil {
					if e2, err := atoi(expr2.(string)); err == nil {
						conditions = append(conditions, e1 < e2)
					} else {
						log.Fatalf("2nd expression in operation not a number.")
					}
				} else {
					log.Fatalf("1st expression in operation not a number.")
				}
			case "GREATER":
				if e1, err := atoi(expr1.(string)); err == nil {
					if e2, err := atoi(expr2.(string)); err == nil {
						conditions = append(conditions, e1 > e2)
					} else {
						log.Fatalf("2nd expression in operation not a number.")
					}
				} else {
					log.Fatalf("1st expression in operation not a number.")
				}
			default:
				log.Fatalf("%s is not a comparison operator", c.Operator)
			}
		}

		for _, condition := range conditions {
			if condition == true {
				return strconv.FormatBool(true)
			}
		}

		return strconv.FormatBool(false)
	case Add:
		left, _ := atoi(eval_expr(e.Left, variables, functions).(string))
		right, _ := atoi(eval_expr(e.Right, variables, functions).(string))
		return strconv.Itoa(left + right)
	case Sub:
		left, _ := atoi(eval_expr(e.Left, variables, functions).(string))
		right, _ := atoi(eval_expr(e.Right, variables, functions).(string))
		return strconv.Itoa(left - right)
	case Mul:
		left, _ := atoi(eval_expr(e.Left, variables, functions).(string))
		right, _ := atoi(eval_expr(e.Right, variables, functions).(string))
		return strconv.Itoa(left * right)
	case FuncCallExpr:
		return func_call_and_return(e, variables, functions)
	default:
		log.Fatalf("Unknown expression %s of type %s", expr, reflect.TypeOf(expr).String())
	}
	return ""
}

// Main

// embed stuff --------------------------------
//
//go:embed examples/installer.yz
var script string

// end embed stuff ----------------------------

func run_program(source string, verbose *bool) {
	tokens := lexer(source, verbose)
	variables := make(map[string]YZVariable)
	functions := make(map[string]Function)
	parser := new_parser(tokens, variables, functions)
	program := parser.parse()
	run(&program)
}

func main() {
	verbose := flag.Bool("v", false, "Verbose mode enabled? (true or false) (not required)")
	flag.Parse()

	if flag.NArg() < 1 {
		if false == true {
			verbose := false
			run_program(script, &verbose)
			os.Exit(0)
		} else {
			log.Fatalf("Usage: [OPTIONS] FILE")
		}
	}

	file := flag.Arg(0)

	fmt.Print("YZ interpeter Output:\n\n")

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	content := string(data)

	run_program(content, verbose)
}
