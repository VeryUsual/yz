// SPDX-FileCopyrightText: 2026 ark
// SPDX-License-Identifier: GPL-3.0-or-later
//
// The YZ Interpreter
// Licensed under the GNU General Public License v3.0
//

package main

// Imports

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"unicode"
)

// AST Nodes

type Num struct {
	Value int
}

type Str struct {
	Value string
}

type Var struct {
	Name  string
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

type IfStmt struct {
	Condition any
	Then      []any
	Else      []any
}

type Function struct {
	Name       string
	Parameters map[string]string
	Contents   []any
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
	variables map[string]string
	functions map[string]Function
}

type Return struct {
	Value any
}

// Tokenizer

type Token struct {
	Type  string
	Value string
}

func lexer(src string, verbose *bool) []Token {
	var i int = 0
	var tokens []Token = []Token{}

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
			switch word {
			case "let":
				tokens = append(tokens, Token{"LET", word})
			case "println":
				tokens = append(tokens, Token{"PRINTLN", word})
			case "if":
				tokens = append(tokens, Token{"IF", word})
			case "func":
				tokens = append(tokens, Token{"FUNC", word})
			case "return":
				tokens = append(tokens, Token{"RETURN", word})
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
		} else if c == '=' && src[i+1] == '=' {
			tokens = append(tokens, Token{"DOUBLE_EQUAL", string(src[i:i+1])})
			i += 2
		} else if c == '>' {
			if src[i+1] == '=' {
				tokens = append(tokens, Token{"GREATER_EQUAL", string(src[i:i+1])})
				i += 2
			} else {
				tokens = append(tokens, Token{"GREATER", string(c)})
				i += 1
			}
		} else if c == '<' {
			if src[i+1] == '=' {
				tokens = append(tokens, Token{"LESS_THAN_EQUAL", string(src[i:i+1])})
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
	tokens []Token
	pos    int
}

func new_parser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
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
	return Program{stmts, make(map[string]string), make(map[string]Function)}
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
	} else if p.cur().Type == "IDENT" {
		if p.peek_next().Type == "LPAREN" {
			return p.func_call_statement()
		} else {
			return p.expr()
		}
	} else {
		log.Fatalf("Unexpected statement token: %s", p.cur().Type)
		os.Exit(0)
		return 0
	}
}

func (p *Parser) let_statement() Let {
	p.eat("LET")
	name := p.eat("IDENT").Value
	p.eat("EQUAL")
	value := p.expr()
	p.eat("SEMI")
	return Let{name, value}
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
	p.eat("IF")
	expr1 := eval_expr(p.expr(), make(map[string]string), make(map[string]Function))
	comparison_operator := Token{"null", "null"}
	switch p.cur().Type {
	case "DOUBLE_EQUAL", "LESS_THAN_EQUAL", "LESS_THAN", "GREATER_EQUAL", "GREATER":
		comparison_operator = p.eat(p.cur().Type)
	default:
		log.Fatalf("%s is not a comparison operator", p.cur().Type)
	}
	expr2 := eval_expr(p.expr(), make(map[string]string), make(map[string]Function))
	p.eat("LBRACE")

	condition := false
	switch comparison_operator.Type {
	case "DOUBLE_EQUAL":
		condition = expr1 == expr2
	case "LESS_THAN_EQUAL":
		if e1, err := strconv.Atoi(expr1); err == nil {
			if e2, err := strconv.Atoi(expr2); err == nil {
				condition = e1 <= e2
			} else {
				log.Fatalf("2nd expression in operation not a number.")
			}
		} else {
			log.Fatalf("1st expression in operation not a number.")
		}
	case "GREATER_EQUAL":
		if e1, err := strconv.Atoi(expr1); err == nil {
			if e2, err := strconv.Atoi(expr2); err == nil {
				condition = e1 >= e2
			} else {
				log.Fatalf("2nd expression in operation not a number.")
			}
		} else {
			log.Fatalf("1st expression in operation not a number.")
		}
	case "LESS_THAN":
		if e1, err := strconv.Atoi(expr1); err == nil {
			if e2, err := strconv.Atoi(expr2); err == nil {
				condition = e1 < e2
			} else {
				log.Fatalf("2nd expression in operation not a number.")
			}
		} else {
			log.Fatalf("1st expression in operation not a number.")
		}
	case "GREATER":
		if e1, err := strconv.Atoi(expr1); err == nil {
			if e2, err := strconv.Atoi(expr2); err == nil {
				condition = e1 > e2
			} else {
				log.Fatalf("2nd expression in operation not a number.")
			}
		} else {
			log.Fatalf("1st expression in operation not a number.")
		}
	default:
		log.Fatalf("%s is not a comparison operator", p.cur().Type)
	}

	thenStmts := []any{}
	for p.cur().Type != "RBRACE" {
		thenStmts = append(thenStmts, p.statement())
	}
	p.eat("RBRACE")
	return IfStmt{condition, thenStmts, []any{}}
}

func (p *Parser) func_statement() Function {
	args := map[string]string{}

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
	p.eat("LBRACE")
	funcStmts := []any{}
	for p.cur().Type != "RBRACE" {
		funcStmts = append(funcStmts, p.statement())
	}
	p.eat("RBRACE")
	return Function{name, args, funcStmts}
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

func func_call_and_return(call FuncCallExpr, variables map[string]string, functions map[string]Function) string {
	if fn, exists := functions[call.Name]; exists {
		func_vars := make(map[string]string)
        for key, value := range variables {
            func_vars[key] = value
        }
        for param, param_value := range call.Parameters {
            if _, ok := fn.Parameters[param]; ok {
                func_vars[param] = eval_expr(param_value, variables, functions)
            } else {
                if fn.Parameters["_yz_arbitrary_params_allowed_"] == "YES" {
                    func_vars[param] = eval_expr(param_value, variables, functions)
                } else {
                    log.Fatalf("Call to function %s failed due to non-existent parameter %s without _yz_arbitrary_params_allowed_ flag.", call.Name, param)
                }
            }
        }
        for name := range fn.Parameters {
            if name != "_yz_arbitrary_params_allowed_" {
                if _, ok := call.Parameters[name]; !ok {
                    log.Fatalf("Missing required parameter %s", name)
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
			val, _ := strconv.Atoi(tok.Value)
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
		run_statement(stmt, program.variables, program.functions)
	}
}

func run_statement(stmt any, variables map[string]string, functions map[string]Function) string {
	switch s := stmt.(type) {
	case Let:
		value := eval_expr(s.Value, variables, functions)
		variables[s.Name] = value
		return ""
	case Print:
		value := eval_expr(s.Expr, variables, functions)
		fmt.Println(value)
		return ""
	case IfStmt:
		if s.Condition == true {
			for _, thenStmt := range s.Then {
				run_statement(thenStmt, variables, functions)
			}	
		}
		return ""
	case Function:
		functions[s.Name] = s
		return ""
	case FuncCallStatement:
		if fn, exists := functions[s.Name]; exists {
			func_vars := make(map[string]string)

			for key, value := range variables {
				func_vars[key] = value
			}

			for param, param_value := range s.Parameters {
				if _, ok := fn.Parameters[param]; ok {
					func_vars[param] = eval_expr(param_value, variables, functions)
				} else {
					if functions[s.Name].Parameters["_yz_arbitrary_params_allowed_"] == "YES" {
						func_vars[param] = eval_expr(param_value, variables, functions)
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
				run_statement(stmt, func_vars, functions)
			}
		} else {
			log.Fatalf("Call to non-existent function %s.", s.Name)
		}
		return ""
	case Return:
		return eval_expr(s.Value, variables, functions)
	default:
		log.Fatalf("Unknown statement:\nType: %s\nValue: %s\n\n", reflect.TypeOf(s).String(), s)
		return ""
	}
}

func eval_expr(expr any, variables map[string]string, functions map[string]Function) string {
	switch e := expr.(type) {
	case Num:
		return strconv.Itoa(e.Value)
	case Str:
		return e.Value
	case Var:
		if _, ok := variables[e.Name]; ok {
			return variables[e.Name]
		} else {
			log.Fatalf("Reference of non-existent variable %s", e.Name)
		}
	case Add:
		left, _ := strconv.Atoi(eval_expr(e.Left, variables, functions))
		right, _ := strconv.Atoi(eval_expr(e.Right, variables, functions))
		return strconv.Itoa(left + right)
	case Sub:
		left, _ := strconv.Atoi(eval_expr(e.Left, variables, functions))
		right, _ := strconv.Atoi(eval_expr(e.Right, variables, functions))
		return strconv.Itoa(left - right)
	case Mul:
		left, _ := strconv.Atoi(eval_expr(e.Left, variables, functions))
		right, _ := strconv.Atoi(eval_expr(e.Right, variables, functions))
		return strconv.Itoa(left * right)
	case FuncCallExpr:
		return func_call_and_return(e, variables, functions)
	default:
		log.Fatalf("Unknown expression %s of type %s", expr, reflect.TypeOf(expr).String())
	}
	return ""
}

// Main

func run_program(source string, verbose *bool) {
	tokens := lexer(source, verbose)
	parser := new_parser(tokens)
	program := parser.parse()
	run(&program)
}

func main() {
	verbose := flag.Bool("v", false, "Verbose mode enabled? (true or false) (not required)")
	flag.Parse()

	fmt.Print("YZ interpeter Output:\n\n")

	data, err := os.ReadFile("examples/0.yz")
	if err != nil {
		log.Fatal(err)
	}

	content := string(data)

	run_program(content, verbose)
}
