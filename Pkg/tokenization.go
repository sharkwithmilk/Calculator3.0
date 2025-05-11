package pkg

import (
	"fmt"
	"strconv"
	"unicode"
)

// Типы токенов
const (
	TokenNumber = iota
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenLParen
	TokenRParen
)

// Token описывает лексему (число, оператор или скобку).
type Token struct {
	Type  int
	Value string
}

// tokenize разбивает исходную строку на отдельные лексемы.
func Tokenize(input string) []Token {
	var tokens []Token

	fmt.Printf("[Токенизация] Исходная строка: %s\n", input) // Отладка: вывод исходной строки

	for idx := 0; idx < len(input); {
		ch := input[idx]

		// Пропускаем пробельные символы
		if ch == ' ' {
			idx++
			continue
		}

		// Обработка чисел (целых и дробных)
		if unicode.IsDigit(rune(ch)) || ch == '.' {
			startIdx := idx
			for idx < len(input) && (unicode.IsDigit(rune(input[idx])) || input[idx] == '.') {
				idx++
			}
			token := Token{Type: TokenNumber, Value: input[startIdx:idx]}
			tokens = append(tokens, token)
			fmt.Printf("[Токенизация] Числовая лексема: %s\n", token.Value) // Отладка: найдено число
			continue
		}

		// Обработка операторов и скобок
		switch ch {
		case '+', '-', '*', '/', '(', ')':
			ttype := map[byte]int{
				'+': TokenPlus, '-': TokenMinus, '*': TokenMultiply,
				'/': TokenDivide, '(': TokenLParen, ')': TokenRParen,
			}[ch]
			tokens = append(tokens, Token{Type: ttype, Value: string(ch)})
			fmt.Printf("[Токенизация] Оператор или скобка: %c\n", ch) // Отладка: найден оператор или скобка
		default:
			fmt.Printf("[Ошибка] Неизвестный символ: '%c' (позиция %d)\n", ch, idx) // Отладка: неизвестный символ
		}
		idx++
	}

	fmt.Printf("[Токенизация] Финальный список токенов: %+v\n", tokens) // Отладка: вывод списка токенов
	return tokens
}

// Node представляет узел AST.
type Node struct {
	Operator string  // Оператор: "+", "-", "*", "/" или пустая строка для числовых узлов.
	Left     *Node   // Левый операнд
	Right    *Node   // Правый операнд
	Value    float64 // Числовое значение, если узел представляет число
}

// Parser содержит набор токенов и текущую позицию разбора.
type Parser struct {
	Tokens []Token
	pos    int
}

// Current возвращает текущий токен разбора.
func (p *Parser) Current() Token {
	if p.pos < len(p.Tokens) {
		return p.Tokens[p.pos]
	}
	return Token{Type: -1} // Индикатор конца потока токенов
}

// Eat принимает токен указанного типа и переходит к следующему.
func (p *Parser) Eat(tokenType int) Token {
	curToken := p.Current()
	if curToken.Type == tokenType {
		p.pos++
		fmt.Printf("[Парсер] Принят токен: %+v\n", curToken) // Отладка: принят токен
		return curToken
	}
	panic(fmt.Sprintf("[Ошибка] Ожидался токен типа %d, получен: %v\n", tokenType, curToken))
}

// ParseFactor обрабатывает число или выражение, заключённое в скобки.
func (p *Parser) ParseFactor() *Node {
	curToken := p.Current()

	if curToken.Type == TokenNumber {
		p.Eat(TokenNumber)
		val, err := strconv.ParseFloat(curToken.Value, 64)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[Парсер (число/скобка)] Создан числовой узел: %v\n", val) // Отладка: создан числовой узел
		return &Node{Value: val}
	} else if curToken.Type == TokenLParen {
		p.Eat(TokenLParen)
		node := p.ParseExpression()
		p.Eat(TokenRParen)
		return node
	}

	panic("[Ошибка] Ожидалось число или '(', получено иное значение")
}

// ParseTerm обрабатывает операции умножения и деления.
func (p *Parser) ParseTerm() *Node {
	result := p.ParseFactor()

	for {
		curToken := p.Current()
		if curToken.Type == TokenMultiply || curToken.Type == TokenDivide {
			p.Eat(curToken.Type)
			rightNode := p.ParseFactor()
			fmt.Printf("[Парсер (умножение/деление)] Выполняется операция: %s с операндами (%v, %v)\n", curToken.Value, result, rightNode) // Отладка: операция умножения или деления
			if curToken.Type == TokenMultiply {
				result = &Node{Operator: "*", Left: result, Right: rightNode, Value: result.Value * rightNode.Value}
			} else {
				result = &Node{Operator: "/", Left: result, Right: rightNode, Value: result.Value / rightNode.Value}
			}
		} else {
			break
		}
	}

	return result
}

// ParseExpression обрабатывает операции сложения и вычитания.
func (p *Parser) ParseExpression() *Node {
	result := p.ParseTerm()

	for {
		curToken := p.Current()
		if curToken.Type == TokenPlus || curToken.Type == TokenMinus {
			p.Eat(curToken.Type)
			rightNode := p.ParseTerm()
			fmt.Printf("[Парсер (сложение/вычитание)] Выполняется операция: %s с операндами (%v, %v)\n", curToken.Value, result, rightNode) // Отладка: операция сложения или вычитания
			if curToken.Type == TokenPlus {
				result = &Node{Operator: "+", Left: result, Right: rightNode, Value: result.Value + rightNode.Value}
			} else {
				result = &Node{Operator: "-", Left: result, Right: rightNode, Value: result.Value - rightNode.Value}
			}
		} else {
			break
		}
	}

	return result
}

// PrintAST возвращает строковое представление AST для отладки.
func PrintAST(node *Node) string {
	if node == nil {
		return ""
	}
	if node.Operator == "" {
		return fmt.Sprintf("%v", node.Value)
	}
	return fmt.Sprintf("(%s %s %s)", PrintAST(node.Left), node.Operator, PrintAST(node.Right))
}
