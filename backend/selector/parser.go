package selector

import "fmt"

type Operator string

const (
	// doubleEqualSignOpeator represents ==
	DoubleEqualSignOperator Operator = "=="
	// notEqualOperator represents !=
	NotEqualOperator Operator = "!="
	// inOperator represents in
	InOperator Operator = "in"
	// notInOperator represents notin
	NotInOperator Operator = "notin"
	// matchesOperator represents matches
	MatchesOperator Operator = "matches"
)

type OperationType int

const (
	OperationTypeFieldSelector OperationType = 0
	OperationTypeLabelSelector OperationType = 1
)

// parser evaluates the tokens produced by the lexer from the input string
type parser struct {
	lexer    *lexer
	position int
	results  []Token
}

// Operation represents a computation, operation, on an LValue and a set of
// RValues.
type Operation struct {
	LValue        string
	Operator      Operator
	RValues       []string
	OperationType OperationType
}

// Parse is deprecated. Use ParseFieldSelector or ParseLabelSelector.
func Parse(input string) (*Selector, error) {
	parser := &parser{lexer: newLexer(input)}

	parser.tokenize()

	operations, err := parser.operations()
	if err != nil {
		return nil, err
	}

	selector := &Selector{Operations: operations}

	return selector, nil
}

// ParseFieldSelector parses the input and returns a field selector.
func ParseFieldSelector(input string) (*Selector, error) {
	sel, err := Parse(input)
	if err != nil {
		return nil, err
	}
	for i := range sel.Operations {
		sel.Operations[i].OperationType = OperationTypeFieldSelector
	}
	return sel, nil
}

// ParseLabelSelector parses the input and returns a label selector.
func ParseLabelSelector(input string) (*Selector, error) {
	sel, err := Parse(input)
	if err != nil {
		return nil, err
	}
	for i := range sel.Operations {
		sel.Operations[i].OperationType = OperationTypeLabelSelector
	}
	return sel, nil
}

// backtrack returns the position to its original place before the last read
// occurred
func (p *parser) backtrack() {
	if p.position >= 1 {
		p.position--
	} else {
		p.position = 0
	}
}

func (p *parser) parseOperator() (Operator, error) {
	result := p.read()
	switch result.Type {
	case doubleEqualSignToken:
		return DoubleEqualSignOperator, nil
	case notEqualToken:
		return NotEqualOperator, nil
	case inToken:
		return InOperator, nil
	case notInToken:
		return NotInOperator, nil
	case matchesToken:
		return MatchesOperator, nil
	default:
		return "", fmt.Errorf("unexpected operator '%s' found", result.Value)
	}
}

// parseOperation analyzes the next results to determine the next operation
func (p *parser) parseOperation() (Operation, error) {
	var r Operation

	// First identify the key
	result := p.read()
	r.LValue = result.Value

	// Now identify the operator
	var err error
	r.Operator, err = p.parseOperator()
	if err != nil {
		return r, err
	}

	// Finally, identify the value
	switch r.Operator {
	case InOperator, NotInOperator:
		r.RValues, err = p.parseValues()
		if err != nil {
			return r, err
		}
	default:
		result := p.read()
		switch result.Type {
		case identifierToken, stringToken, boolToken, matchesToken:
			r.RValues = []string{result.Value}
		default:
			return r, fmt.Errorf("unexpected token '%s': expected an identifier or literal value", result.Value)
		}
	}

	return r, nil
}

// parseValues parses values found in an array used by the 'in' & 'notin'
// operators, e.g. (a,b,c)
func (p *parser) parseValues() ([]string, error) {
	var values []string
	// The first token should be '[' or a selector
	result := p.read()
	if result.Type == identifierToken {
		return []string{result.Value}, nil
	}
	if result.Type != leftSquareToken {
		return values, fmt.Errorf("found '%s', expected '['", result.Value)
	}

	for {
		result = p.read()
		switch result.Type {
		case identifierToken, stringToken:
			values = append(values, result.Value)
		case commaToken:
			continue
		case rightSquareToken:
			return values, nil
		default:
			return values, fmt.Errorf("unexpected token '%s', expected a comma or an identifier", result.Value)
		}
	}
}

// peek returns the next result (token and its concrete value) without consuming
// it by not moving the position
func (p *parser) peek() Token {
	defer p.backtrack()
	return p.read()
}

// read returns the next rune in the input and consumes it by moving the
// position
func (p *parser) read() Token {
	p.position++
	return p.results[p.position-1]
}

// operations analyzes the results and determines the list of operations
func (p *parser) operations() ([]Operation, error) {
	var operations []Operation

	for {
		result := p.peek()
		switch result.Type {
		case identifierToken, stringToken:
			operation, err := p.parseOperation()
			if err != nil {
				return nil, fmt.Errorf("could not parse the operations: %s", err)
			}
			// We found a valid operation, append it to our list of operations
			operations = append(operations, operation)
		case doubleAmpersandToken:
			// Move the position forward
			_ = p.read()

			// Make sure we already have a operation before accepting the '&&'
			// operator
			if len(operations) == 0 {
				return operations, fmt.Errorf("unexpected '&&' operator found, expected a operation first")
			}

			// Make sure the next token is an identifier or a string, which are
			// the 2 types of tokens that can start a new expression.
			result = p.peek()
			if result.Type == identifierToken || result.Type == stringToken {
				continue
			} else {
				return nil, fmt.Errorf("unexpected token '%s', expected an identifier or string after '&&'", result.Value)
			}
		case endOfStringToken:
			return operations, nil
		default:
			return nil, fmt.Errorf("unexpected token '%s', expected an identifier or end of string", result.Value)
		}

	}
}

// tokenize goes through the input string and produces a list of tokens stored
// into the results attribute
func (p *parser) tokenize() {
	var token Token
	for token.Type != endOfStringToken {
		token = p.lexer.Tokenize()
		p.results = append(p.results, token)
	}
}
