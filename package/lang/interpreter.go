package lang

import (
	"fmt"
)

type interpreter struct {
	stackTop *stackFrame
}

type Caller interface {
	Call(vFunction, []Value) (Value, error)
}

// TODO: callLookuper interface
// to pass in places?

func NewInterpreter(rootScope *Scope, expr Expr) *interpreter {
	return &interpreter{
		stackTop: &stackFrame{
			expr:  expr,
			scope: rootScope,
		},
	}
}

func (i *interpreter) Interpret() (Value, error) {
	return i.stackTop.expr.Evaluate(i)
}

func (i *interpreter) pushFrame(frame *stackFrame) {
	frame.parentFrame = i.stackTop
	i.stackTop = frame
}

func (i *interpreter) popFrame() *stackFrame {
	if i.stackTop == nil {
		panic("can'out pop frame; at bottom")
	}
	top := i.stackTop
	i.stackTop = top.parentFrame
	return top
}

func (i *interpreter) Call(vFunc vFunction, argVals []Value) (Value, error) {
	// Make new scope.
	newScope := NewScope(i.stackTop.scope)
	params := vFunc.GetParamList()
	if len(params) != len(argVals) {
		// Checked when we get the type.
		panic("wrong number of args")
	}
	for idx, argVal := range argVals {
		param := params[idx]
		newScope.Add(param.Name, argVal)
	}
	// Make and push new stack frame.
	newFrame := &stackFrame{
		scope: newScope,
		vFunc: vFunc,
	}
	i.pushFrame(newFrame)
	// Call the lambda or builtin.
	var val Value
	var err error
	switch tVFunc := vFunc.(type) {
	case *vLambda:
		newFrame.expr = tVFunc.def.body
		val, err = i.Interpret()
		return val, err
	case *VBuiltin:
		val, err = tVFunc.Impl(i, argVals)
	}
	// Pop and return.
	i.popFrame()
	return val, err
}

type Scope struct {
	parent *Scope
	vals   map[string]Value
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		vals:   map[string]Value{},
		parent: parent,
	}
}

type notInScopeError struct {
	name string
}

func (e *notInScopeError) Error() string {
	return fmt.Sprintf("not in scope: %s", e.name)
}

func (s *Scope) find(name string) (Value, error) {
	val, ok := s.vals[name]
	if !ok {
		if s.parent != nil {
			return s.parent.find(name)
		}
		return nil, &notInScopeError{
			name: name,
		}
	}
	return val, nil
}

func (s *Scope) Add(name string, value Value) {
	s.vals[name] = value
}

type stackFrame struct {
	// if parentFrame is null, this is the root frame.
	parentFrame *stackFrame
	expr        Expr
	scope       *Scope

	// if it's a function stack frame
	vFunc vFunction
	// if it's a object key stack frame
	objKey string
	// if it's a record stack frame
	primaryKey Value
}

// TODO: stack frame and stuff
// keep the func name in there
// also keep a query path of some kind in there,
// so we can go back up the stack and install live query
// listeners.

func Interpret(e Expr, rootScope *Scope) (Value, error) {
	i := NewInterpreter(rootScope, e)
	return i.Interpret()
}