package lang

import (
	"fmt"

	pp "github.com/vilterp/treesql/package/pretty_print"
)

type Expr interface {
	Evaluate(*interpreter) (Value, error)

	GetType(*Scope) (Type, error)

	Format() pp.Doc
}

// Int

type EIntLit int

var eZero = EIntLit(9)
var _ Expr = &eZero

// TODO: can we avoid an allocation here?
func (e *EIntLit) Evaluate(_ *interpreter) (Value, error) {
	theInt := VInt(*e)
	return &theInt, nil
}

func (e *EIntLit) Format() pp.Doc {
	return pp.Textf("%d", *e)
}

func (e *EIntLit) GetType(_ *Scope) (Type, error) {
	return TInt, nil
}

// String

type EStringLit string

var eEmptyStr = EStringLit("")
var _ Expr = &eEmptyStr

func (e *EStringLit) Evaluate(_ *interpreter) (Value, error) {
	theStr := VString(*e)
	return &theStr, nil
}

func (e *EStringLit) Format() pp.Doc {
	return pp.Textf("%#v", *e)
}

func (e *EStringLit) GetType(_ *Scope) (Type, error) {
	return TString, nil
}

// Var

type EVar struct {
	name string
}

var _ Expr = &EVar{}

func (e *EVar) Evaluate(interp *interpreter) (Value, error) {
	return interp.stackTop.scope.find(e.name)
}

func (e *EVar) Format() pp.Doc {
	return pp.Text(e.name)
}

func (e *EVar) GetType(scope *Scope) (Type, error) {
	val, err := scope.find(e.name)
	if err != nil {
		return nil, err
	}
	return val.GetType(), nil
}

// Object

type EObjectLit struct {
	exprs map[string]Expr
}

var _ Expr = &EObjectLit{}

func (ol *EObjectLit) Evaluate(interp *interpreter) (Value, error) {
	// TODO: push an object path frame
	vals := map[string]Value{}

	for name, expr := range ol.exprs {
		val, err := expr.Evaluate(interp)
		if err != nil {
			return nil, err
		}
		vals[name] = val
	}

	return &VObject{
		vals: vals,
	}, nil
}

func (ol *EObjectLit) Format() pp.Doc {
	// Sort keys
	keys := make([]string, len(ol.exprs))
	idx := 0
	for k := range ol.exprs {
		keys[idx] = k
		idx++
	}

	kvDocs := make([]pp.Doc, len(ol.exprs))
	for idx, key := range keys {
		kvDocs[idx] = pp.Concat([]pp.Doc{
			pp.Text(key),
			pp.Text(": "),
			ol.exprs[key].Format(),
		})
	}

	return pp.Concat([]pp.Doc{
		pp.Text("("), pp.Newline,
		pp.Nest(pp.Concat(kvDocs), 2),
		pp.Text("}"), pp.Newline,
	})
}

func (ol *EObjectLit) GetType(scope *Scope) (Type, error) {
	types := map[string]Type{}

	for name, expr := range ol.exprs {
		typ, err := expr.GetType(scope)
		if err != nil {
			return nil, err
		}
		types[name] = typ
	}

	return &TObject{
		Types: types,
	}, nil
}

// Lambda

type Param struct {
	Name string
	Typ  Type
}

type ELambda struct {
	params  ParamList
	body    Expr
	retType Type
}

var _ Expr = &ELambda{}

func (l *ELambda) Evaluate(interp *interpreter) (Value, error) {
	return &vLambda{
		def: l,
		// TODO: don't close over the scope if we don't need anything from there
		definedInScope: interp.stackTop.scope,
	}, nil
}

func (l *ELambda) Format() pp.Doc {
	// TODO: use concat instead of sprintf here to facilitate line breaking
	// when it has that
	return pp.Text(
		fmt.Sprintf(
			"(%s): %s => (%s)",
			l.params.Format().Render(), l.retType.Format().Render(), l.body.Format().Render(),
		),
	)
}

func (l *ELambda) GetType(_ *Scope) (Type, error) {
	return &tFunction{
		params:  l.params,
		retType: l.retType,
	}, nil
}

// Func call

type EFuncCall struct {
	funcName string
	args     []Expr
}

// TODO: may need to pass interpreter in here
// so this can push a stack frame
// and interpret the inner expr
func (fc *EFuncCall) Evaluate(interp *interpreter) (Value, error) {
	// Get function value.
	funcVal, err := interp.stackTop.scope.find(fc.funcName)
	if err != nil {
		return nil, err
	}

	// Get argument values.
	argVals := make([]Value, len(fc.args))
	for idx, argExpr := range fc.args {
		argVal, err := argExpr.Evaluate(interp)
		if err != nil {
			return nil, err
		}
		argVals[idx] = argVal
	}

	switch tFuncVal := funcVal.(type) {
	case *vLambda:
		return interp.call(tFuncVal, argVals)
	case *vBuiltin:
		return interp.call(tFuncVal, argVals)
	default:
		return nil, fmt.Errorf("not a function: %s", fc.funcName)
	}
}

func (fc *EFuncCall) Format() pp.Doc {
	argDocs := make([]pp.Doc, len(fc.args))
	for idx, arg := range fc.args {
		argDocs[idx] = arg.Format()
	}

	return pp.Concat([]pp.Doc{
		pp.Text(fc.funcName),
		pp.Text("("),
		pp.Join(argDocs, pp.Text(", ")),
		pp.Text(")"),
	})
}

func (fc *EFuncCall) GetType(scope *Scope) (Type, error) {
	funcVal, err := scope.find(fc.funcName)
	if err != nil {
		return nil, err
	}
	switch tFuncVal := funcVal.(type) {
	case vFunction:
		return tFuncVal.GetRetType(), nil
	default:
		return nil, fmt.Errorf("not a function: %s", fc.funcName)
	}
}

// TODO: Let binding
