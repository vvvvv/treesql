package lang

var BuiltinsScope *Scope
var BuiltinsTypeScope *TypeScope

func init() {
	BuiltinsScope = NewScope(nil)
	BuiltinsScope.Add("plus", &VBuiltin{
		Name:    "plus",
		Params:  []Param{{"a", TInt}, {"b", TInt}},
		RetType: TInt,
		Impl: func(_ Caller, args []Value) (Value, error) {
			l := int(*mustBeVInt(args[0]))
			r := int(*mustBeVInt(args[1]))
			return NewVInt(l + r), nil
		},
	})
	BuiltinsScope.Add("map", &VBuiltin{
		Name: "map",
		Params: []Param{
			{"iter", NewTIterator(NewTVar("A"))},
			{"func", &tFunction{
				params:  []Param{{"x", NewTVar("A")}},
				retType: NewTVar("B"),
			}},
		},
		RetType: NewTIterator(NewTVar("B")),
		Impl: func(c Caller, args []Value) (Value, error) {
			f := mustBeVFunction(args[1])
			return &VIteratorRef{
				iterator: &mapIterator{
					innerIterator: mustBeVIteratorRef(args[0]).iterator,
					f:             f,
				},
				ofType: f.GetRetType(),
			}, nil
		},
	})

	BuiltinsTypeScope = BuiltinsScope.toTypeScope()
}

// TODO:
// comparision
// arithmetic
// maybe record subset and update