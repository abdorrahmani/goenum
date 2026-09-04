package generator

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"strconv"
)

// evalExpr evaluates the subset of constant expressions the generator
// supports: literals, previously-defined enum constants, iota, parentheses,
// unary -/+/^ and the binary arithmetic/bitwise operators. Anything else is
// rejected with a positioned error so unsupported declarations fail at
// generation time instead of producing wrong tables.
func evalExpr(fset *token.FileSet, expr ast.Expr, iotaVal int64, consts map[string]constant.Value) (constant.Value, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return constant.MakeFromLiteral(e.Value, e.Kind, 0), nil
	case *ast.Ident:
		switch {
		case e.Name == "iota":
			return constant.MakeInt64(iotaVal), nil
		case e.Name == "true":
			return constant.MakeBool(true), nil
		case e.Name == "false":
			return constant.MakeBool(false), nil
		}
		if v, ok := consts[e.Name]; ok {
			return v, nil
		}
		return nil, posErr(fset, e.Pos(), "", fmt.Sprintf("unsupported identifier %q in const expression (only literals, iota and earlier constants of the same enum are allowed)", e.Name))
	case *ast.ParenExpr:
		return evalExpr(fset, e.X, iotaVal, consts)
	case *ast.UnaryExpr:
		x, err := evalExpr(fset, e.X, iotaVal, consts)
		if err != nil {
			return nil, err
		}
		switch e.Op {
		case token.SUB:
			return constant.UnaryOp(token.SUB, x, 0), nil
		case token.XOR:
			return constant.UnaryOp(token.XOR, x, 0), nil
		case token.ADD:
			return x, nil
		}
		return nil, posErr(fset, e.Pos(), "", fmt.Sprintf("unsupported unary operator %v", e.Op))
	case *ast.BinaryExpr:
		x, err := evalExpr(fset, e.X, iotaVal, consts)
		if err != nil {
			return nil, err
		}
		y, err := evalExpr(fset, e.Y, iotaVal, consts)
		if err != nil {
			return nil, err
		}
		switch e.Op {
		case token.SHL, token.SHR:
			// go/constant requires an unsigned shift count; do the shift
			// directly on the integer value instead.
			n, ok := constant.Uint64Val(y)
			if !ok || n >= 64 {
				return nil, posErr(fset, e.Pos(), "", "shift count must be an unsigned integer constant below 64")
			}
			if x.Kind() != constant.Int {
				return nil, posErr(fset, e.Pos(), "", "shift operand must be an integer constant")
			}
			if i, ok := constant.Int64Val(x); ok {
				if e.Op == token.SHL {
					return constant.MakeInt64(i << n), nil
				}
				return constant.MakeInt64(i >> n), nil
			}
			if u, ok := constant.Uint64Val(x); ok {
				if e.Op == token.SHL {
					return constant.MakeUint64(u << n), nil
				}
				return constant.MakeUint64(u >> n), nil
			}
			return nil, posErr(fset, e.Pos(), "", "shift operand too large")
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
			token.AND, token.OR, token.XOR, token.AND_NOT:
			return constant.BinaryOp(x, e.Op, y), nil
		default:
			return nil, posErr(fset, e.Pos(), "", fmt.Sprintf("unsupported binary operator %v", e.Op))
		}
	}
	return nil, posErr(fset, expr.Pos(), "", fmt.Sprintf("unsupported const expression %s", exprKind(expr)))
}

func exprKind(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.CallExpr:
		return "conversion or call"
	case *ast.SelectorExpr:
		return "selector"
	default:
		return strconv.Quote(fmt.Sprintf("%T", expr))
	}
}
