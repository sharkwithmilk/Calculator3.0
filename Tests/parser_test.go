package Tests

import (
	"Calculator3.0/Pkg"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []pkg.Token
	}{
		{"2 + 3", []pkg.Token{{Type: pkg.TokenNumber, Value: "2"}, {Type: pkg.TokenPlus, Value: "+"}, {Type: pkg.TokenNumber, Value: "3"}}},
		{"(2 * 3)", []pkg.Token{{Type: pkg.TokenLParen, Value: "("}, {Type: pkg.TokenNumber, Value: "2"}, {Type: pkg.TokenMultiply, Value: "*"}, {Type: pkg.TokenNumber, Value: "3"}, {Type: pkg.TokenRParen, Value: ")"}}},
		{"invalid", []pkg.Token{}},
	}

	for _, tt := range tests {
		tokens := pkg.Tokenize(tt.input)
		if len(tokens) != len(tt.expected) {
			t.Errorf("Для '%s' ожидалось %d токенов, получено %d", tt.input, len(tt.expected), len(tokens))
			continue
		}
		for i, token := range tokens {
			if token.Type != tt.expected[i].Type || token.Value != tt.expected[i].Value {
				t.Errorf("Для '%s' ожидался токен %+v, получен %+v", tt.input, tt.expected[i], token)
			}
		}
	}
}

func TestParseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2 + 3", "(2 + 3)"},             // Было "(+ 2 3)"
		{"2 + 3 * 4", "(2 + (3 * 4))"},   // Было "(+ 2 (* 3 4))"
		{"(2 + 3) * 4", "((2 + 3) * 4)"}, // Было "(* (+ 2 3) 4)"
	}

	for _, tt := range tests {
		tokens := pkg.Tokenize(tt.input)
		parser := &pkg.Parser{Tokens: tokens}
		ast := parser.ParseExpression()
		astStr := pkg.PrintAST(ast)
		if astStr != tt.expected {
			t.Errorf("Для '%s' ожидалось AST '%s', получено '%s'", tt.input, tt.expected, astStr)
		}
	}
}
func TestParseInvalidExpression(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Ожидалась паника при парсинге некорректного выражения")
		}
	}()
	tokens := pkg.Tokenize("2 + * 3")
	parser := &pkg.Parser{Tokens: tokens}
	parser.ParseExpression()
}
