package normalizer

import (
	"errors"

	"github.com/bblfsh/python-driver/driver/normalizer/pyast"

	"gopkg.in/bblfsh/sdk.v1/uast"
	. "gopkg.in/bblfsh/sdk.v1/uast/ann"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer/annotatter"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer/positioner"
)

/*
   Tip: to quickly see the native AST generated by Python you can do:
   from ast import *
   code = 'print("test code")'
   dump(parse(code))
*/

/*
uast.For a description of Python AST nodes:

https://greentreesnakes.readthedocs.io/en/latest/nodes.html

*/

/*
Unmarked nodes or nodes needing new features from the SDK:

   These nodes would need a list-mix feature to convert parallel lists
   into list of parents and children:

   BoolOp
   arguments.defaults
   Compare.comparators
   Compare.ops
   uast.Ifuast.Condition.left
	(see: https://greentreesnakes.readthedocs.io/en/latest/nodes.html#Compare)
*/

// Transformers is the of list `transformer.Transfomer` to apply to a UAST, to
// learn more about the Transformers and the available ones take a look to:
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast/transformers
var Transformers = []transformer.Tranformer{
	annotatter.NewAnnotatter(AnnotationRules),
	positioner.NewFillOffsetFromLineCol(),
}

// Common for FunctionDef, AsyncFunctionDef and Lambda
var argumentsAnn = On(pyast.Arguments).Roles(uast.Function, uast.Declaration, uast.Incomplete, uast.Argument).Children(
	On(HasInternalRole("args")).Roles(uast.Function, uast.Declaration, uast.Argument, uast.Name, uast.Identifier),
	On(HasInternalRole("vararg")).Roles(uast.Function, uast.Declaration, uast.Argument, uast.ArgsList, uast.Name, uast.Identifier),
	On(HasInternalRole("kwarg")).Roles(uast.Function, uast.Declaration, uast.Argument, uast.ArgsList, uast.Map, uast.Name, uast.Identifier),
	On(HasInternalRole("kwonlyargs")).Roles(uast.Function, uast.Declaration, uast.Argument, uast.ArgsList, uast.Map, uast.Name, uast.Identifier),
)

// AnnotationRules describes how a UAST should be annotated with `uast.Role`.
//
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast/ann
var AnnotationRules = On(Any).Self(
	On(Not(pyast.Module)).Error(errors.New("root must be uast.Module")),
	On(pyast.Module).Roles(uast.File, uast.Module).Descendants(

		// Binary Expressions
		On(pyast.BinOp).Roles(uast.Expression, uast.Binary).Children(
			On(HasInternalRole("op")).Roles(uast.Expression, uast.Binary, uast.Operator),
			On(HasInternalRole("left")).Roles(uast.Expression, uast.Binary, uast.Left),
			On(HasInternalRole("right")).Roles(uast.Expression, uast.Binary, uast.Right),
		),

		// Comparison operators
		On(pyast.Eq).Roles(uast.Binary, uast.Operator, uast.Equal),
		On(pyast.NotEq).Roles(uast.Binary, uast.Operator, uast.Equal, uast.Not),
		On(pyast.Lt).Roles(uast.Binary, uast.Operator, uast.LessThan),
		On(pyast.LtE).Roles(uast.Binary, uast.Operator, uast.LessThanOrEqual),
		On(pyast.Gt).Roles(uast.Binary, uast.Operator, uast.GreaterThan),
		On(pyast.GtE).Roles(uast.Binary, uast.Operator, uast.GreaterThanOrEqual),
		On(pyast.Is).Roles(uast.Binary, uast.Operator, uast.Identical),
		On(pyast.IsNot).Roles(uast.Binary, uast.Operator, uast.Identical, uast.Not),
		On(pyast.In).Roles(uast.Binary, uast.Operator, uast.Contains),
		On(pyast.NotIn).Roles(uast.Binary, uast.Operator, uast.Contains, uast.Not),

		// Aritmetic operators
		On(pyast.Add).Roles(uast.Binary, uast.Operator, uast.Add),
		On(pyast.Sub).Roles(uast.Binary, uast.Operator, uast.Substract),
		On(pyast.Mult).Roles(uast.Binary, uast.Operator, uast.Multiply),
		On(pyast.Div).Roles(uast.Binary, uast.Operator, uast.Divide),
		On(pyast.Mod).Roles(uast.Binary, uast.Operator, uast.Modulo),
		On(pyast.FloorDiv).Roles(uast.Binary, uast.Operator, uast.Divide, uast.Incomplete),
		On(pyast.Pow).Roles(uast.Binary, uast.Operator, uast.Incomplete),
		On(pyast.MatMult).Roles(uast.Binary, uast.Operator, uast.Multiply, uast.Incomplete),

		// Bitwise operators
		On(pyast.LShift).Roles(uast.Binary, uast.Operator, uast.Bitwise, uast.LeftShift),
		On(pyast.RShift).Roles(uast.Binary, uast.Operator, uast.Bitwise, uast.RightShift),
		On(pyast.BitOr).Roles(uast.Binary, uast.Operator, uast.Bitwise, uast.Or),
		On(pyast.BitXor).Roles(uast.Binary, uast.Operator, uast.Bitwise, uast.Xor),
		On(pyast.BitAnd).Roles(uast.Binary, uast.Operator, uast.Bitwise, uast.And),

		// Boolean operators
		// Not applying the "Binary" role since even while in the Python code
		// boolean operators use (seemingly binary) infix notation, the generated
		// AST nodes use prefix.
		On(pyast.And).Roles(uast.Binary, uast.Operator, uast.Boolean, uast.And),
		On(pyast.Or).Roles(uast.Binary, uast.Operator, uast.Boolean, uast.Or),
		On(pyast.Not).Roles(uast.Binary, uast.Operator, uast.Boolean, uast.Not),
		On(pyast.UnaryOp).Roles(uast.Binary, uast.Operator, uast.Unary, uast.Expression),

		// Unary operators
		On(pyast.Invert).Roles(uast.Operator, uast.Unary, uast.Bitwise, uast.Not),
		On(pyast.UAdd).Roles(uast.Operator, uast.Unary, uast.Positive),
		On(pyast.USub).Roles(uast.Operator, uast.Unary, uast.Negative),

		// Literals
		On(pyast.Str).Roles(uast.Literal, uast.String, uast.Expression, uast.Primitive),
		On(pyast.StringLiteral).Roles(uast.Literal, uast.String, uast.Expression, uast.Primitive),
		On(pyast.Bytes).Roles(uast.Literal, uast.ByteString, uast.Expression, uast.Primitive),
		On(pyast.Num).Roles(uast.Literal, uast.Number, uast.Expression, uast.Primitive).Children(
			On(HasInternalRole("n")).Roles(uast.Literal, uast.Number, uast.Expression),
		),
		On(pyast.BoolLiteral).Roles(uast.Literal, uast.Boolean, uast.Expression, uast.Primitive),
		// another grouping node like "arguments"
		On(pyast.BoolOp).Roles(uast.Expression, uast.Boolean, uast.Incomplete),
		On(pyast.JoinedStr).Roles(uast.Literal, uast.String, uast.Expression, uast.Primitive).Children(
			On(pyast.FormattedValue).Roles(uast.Expression, uast.Incomplete),
		),
		On(pyast.NoneLiteral).Roles(uast.Literal, uast.Null, uast.Expression, uast.Primitive),
		On(pyast.Set).Roles(uast.Literal, uast.Set, uast.Expression, uast.Primitive),
		On(pyast.List).Roles(uast.Literal, uast.List, uast.Expression, uast.Primitive),
		On(pyast.Dict).Roles(uast.Literal, uast.Map, uast.Expression, uast.Primitive).Children(
			On(HasInternalRole("keys")).Roles(uast.Map, uast.Key),
			On(HasInternalRole("values")).Roles(uast.Map, uast.Value),
		),
		On(pyast.Tuple).Roles(uast.Literal, uast.Tuple, uast.Expression, uast.Primitive),

		// FIXME: the FunctionDeclarationReceiver is not set for methods; it should be taken from the parent
		// Type node Token (2 levels up) but the SDK doesn't allow this
		On(pyast.FunctionDef).Roles(uast.Function, uast.Declaration, uast.Name, uast.Identifier).Children(argumentsAnn),
		On(pyast.AsyncFunctionDef).Roles(uast.Function, uast.Declaration, uast.Name, uast.Identifier, uast.Incomplete).Children(argumentsAnn),
		On(pyast.FuncDecorators).Roles(uast.Function, uast.Declaration, uast.Call, uast.Incomplete),
		On(pyast.FuncDefBody).Roles(uast.Function, uast.Declaration, uast.Body),
		// Default arguments: Python's AST puts default arguments on a sibling list to the one of
		// arguments that must be mapped to the arguments right-aligned like:
		// a, b=2, c=3 ->
		//		args    [a,b,c],
		//		defaults  [2,3]
		// TODO: create an issue for the SDK
		On(pyast.ArgumentDefaults).Roles(uast.Function, uast.Declaration, uast.Argument, uast.Value, uast.Incomplete),
		On(pyast.AsyncFuncDecorators).Roles(uast.Function, uast.Declaration, uast.Call, uast.Incomplete),
		On(pyast.AsyncFuncDefBody).Roles(uast.Function, uast.Declaration, uast.Body),
		// FIXME: change to Function, Declaration, ArgumentS once the PR has been merged
		On(pyast.Lambda).Roles(uast.Function, uast.Declaration, uast.Expression, uast.Incomplete).Children(
			On(pyast.LambdaBody).Roles(uast.Function, uast.Declaration, uast.Body),
			argumentsAnn,
		),

		On(pyast.Attribute).Roles(uast.Identifier, uast.Expression).Children(
			On(pyast.Name).Roles(uast.Identifier, uast.Qualified)),

		On(pyast.Call).Roles(uast.Function, uast.Call, uast.Expression).Children(
			On(HasInternalRole("args")).Roles(uast.Function, uast.Call, uast.Positional, uast.Argument, uast.Name),
			On(HasInternalRole("keywords")).Roles(uast.Function, uast.Call, uast.Argument, uast.Name).Children(
				On(HasInternalRole("value")).Roles(uast.Argument, uast.Value),
			),
			On(HasInternalRole("func")).Self(
				On(pyast.Name).Roles(uast.Call, uast.Callee),
				On(pyast.Attribute).Roles(uast.Call, uast.Callee).Children(
					On(HasInternalRole("value")).Roles(uast.Call, uast.Receiver),
				)),
		),

		//
		//	Assign => Assigment:
		//		targets[] => Left
		//		value	  => Right
		//
		On(pyast.Assign).Roles(uast.Binary, uast.Assignment, uast.Expression).Children(
			On(HasInternalRole("targets")).Roles(uast.Left),
			On(HasInternalRole("value")).Roles(uast.Right),
		),

		On(pyast.AugAssign).Roles(uast.Operator, uast.Binary, uast.Assignment, uast.Statement).Children(
			On(HasInternalRole("op")).Roles(uast.Operator, uast.Binary),
			On(HasInternalRole("target")).Roles(uast.Left),
			On(HasInternalRole("value")).Roles(uast.Right),
		),

		On(pyast.Expression).Roles(uast.Expression),
		On(pyast.Expr).Roles(uast.Expression),
		On(pyast.Name).Roles(uast.Identifier, uast.Expression),
		// Comments and non significative whitespace
		On(pyast.SameLineNoops).Roles(uast.Comment),
		On(pyast.PreviousNoops).Roles(uast.Whitespace).Children(
			On(HasInternalRole("lines")).Roles(uast.Comment),
		),
		On(pyast.RemainderNoops).Roles(uast.Whitespace).Children(
			On(HasInternalRole("lines")).Roles(uast.Comment),
		),

		// TODO: check what Constant nodes are generated in the python AST and improve this
		On(pyast.Constant).Roles(uast.Identifier, uast.Expression),
		On(pyast.Try).Roles(uast.Try, uast.Statement).Children(
			On(pyast.TryBody).Roles(uast.Try, uast.Body),
			On(pyast.TryFinalBody).Roles(uast.Try, uast.Finally),
			On(pyast.TryHandlers).Roles(uast.Try, uast.Catch),
			On(pyast.TryElse).Roles(uast.Try, uast.Body, uast.Else),
		),
		On(pyast.TryExcept).Roles(uast.Try, uast.Catch, uast.Statement),     // py2
		On(pyast.ExceptHandler).Roles(uast.Try, uast.Catch, uast.Statement), // py3
		On(pyast.ExceptHandlerName).Roles(uast.Try, uast.Catch, uast.Identifier),
		On(pyast.TryFinally).Roles(uast.Try, uast.Finally, uast.Statement),
		On(pyast.Raise).Roles(uast.Throw, uast.Statement),
		// FIXME: review, add path for the body and items childs
		On(pyast.With).Roles(uast.Block, uast.Scope, uast.Statement),
		On(pyast.WithBody).Roles(uast.Block, uast.Scope, uast.Expression, uast.Incomplete),
		On(pyast.WithItems).Roles(uast.Identifier, uast.Expression, uast.Incomplete),
		On(pyast.AsyncWith).Roles(uast.Block, uast.Scope, uast.Statement, uast.Incomplete),
		On(pyast.Withitem).Roles(uast.Identifier, uast.Incomplete),
		On(pyast.Return).Roles(uast.Return, uast.Statement),
		On(pyast.Break).Roles(uast.Break, uast.Statement),
		On(pyast.Continue).Roles(uast.Continue, uast.Statement),

		// Comparison nodes in Python are oddly structured. Probably one if the first
		// things that could be changed once we can normalize tree structures. Check:
		// https://greentreesnakes.readthedocs.io/en/latest/nodes.html#Compare

		// Parent of all comparisons
		On(pyast.Compare).Roles(uast.Expression, uast.Binary).Children(
			// Operators
			On(pyast.CompareOps).Roles(uast.Expression),
			// Leftmost element (the others are the comparators below)
			On(HasInternalRole("left")).Roles(uast.Expression, uast.Left),
			// These hold the members of the comparison (not the operators)
			On(pyast.CompareComparators).Roles(uast.Expression, uast.Right),
		),
		On(pyast.If).Roles(uast.If, uast.Statement).Children(
			On(pyast.IfBody).Roles(uast.If, uast.Body, uast.Then),
			On(HasInternalRole("test")).Roles(uast.If, uast.Condition),
			On(pyast.IfElse).Roles(uast.If, uast.Body, uast.Else),
		),
		On(pyast.IfExp).Roles(uast.If, uast.Expression).Children(
			// These are used on ifexpressions (a = 1 if x else 2)
			On(HasInternalRole("body")).Roles(uast.If, uast.Body, uast.Then),
			On(HasInternalRole("test")).Roles(uast.If, uast.Condition),
			On(HasInternalRole("orelse")).Roles(uast.If, uast.Body, uast.Else),
		),
		On(pyast.Import).Roles(uast.Import, uast.Declaration, uast.Statement),
		// "y" in "from x import y" or "import y"
		On(pyast.Alias).Roles(uast.Import, uast.Pathname, uast.Identifier),
		// "x" in "from x import y"
		On(pyast.ImportFromModule).Roles(uast.Import, uast.Pathname, uast.Identifier),
		// "y" in "import x as y"
		On(pyast.AliasAsName).Roles(uast.Import, uast.Alias, uast.Identifier),
		On(pyast.ImportFrom).Roles(uast.Import, uast.Declaration, uast.Statement),
		On(pyast.ClassDef).Roles(uast.Type, uast.Declaration, uast.Identifier, uast.Statement).Children(
			On(pyast.ClassDefDecorators).Roles(uast.Type, uast.Call, uast.Incomplete),
			On(pyast.ClassDefBody).Roles(uast.Type, uast.Declaration, uast.Body),
			On(pyast.ClassDefBases).Roles(uast.Type, uast.Declaration, uast.Base),
			On(pyast.ClassDefKeywords).Roles(uast.Incomplete).Children(
				On(pyast.Keyword).Roles(uast.Identifier, uast.Incomplete),
			),
		),

		On(pyast.For).Roles(uast.For, uast.Iterator, uast.Statement).Children(
			On(pyast.ForBody).Roles(uast.For, uast.Body),
			On(HasInternalRole("iter")).Roles(uast.For, uast.Expression),
			On(HasInternalRole("target")).Roles(uast.For, uast.Update),
			On(pyast.ForElse).Roles(uast.For, uast.Body, uast.Else),
		),
		On(pyast.AsyncFor).Roles(uast.For, uast.Iterator, uast.Statement, uast.Incomplete).Children(
			On(pyast.AsyncForBody).Roles(uast.For, uast.Body),
			On(HasInternalRole("iter")).Roles(uast.For, uast.Expression),
			On(HasInternalRole("target")).Roles(uast.For, uast.Update),
			On(pyast.AsyncForElse).Roles(uast.For, uast.Body, uast.Else),
		),
		On(pyast.While).Roles(uast.While, uast.Statement).Children(
			On(pyast.WhileBody).Roles(uast.While, uast.Body),
			On(HasInternalRole("test")).Roles(uast.While, uast.Condition),
			On(pyast.WhileElse).Roles(uast.While, uast.Body, uast.Else),
		),
		On(pyast.Pass).Roles(uast.Noop, uast.Statement),
		On(pyast.Assert).Roles(uast.Assert, uast.Statement),

		// These are AST nodes in Python2 but we convert them to functions in the UAST
		// like they are in Python3
		On(pyast.Exec).Roles(uast.Function, uast.Call, uast.Expression).Children(
			On(HasInternalRole("body")).Roles(uast.Call, uast.Argument, uast.Positional),
			On(HasInternalRole("globals")).Roles(uast.Call, uast.Argument, uast.Positional),
			On(HasInternalRole("locals")).Roles(uast.Call, uast.Argument, uast.Positional),
		),
		// Repr already comes as a uast.Call \o/
		// Print as a function too.
		On(pyast.Print).Roles(uast.Function, uast.Call, uast.Callee, uast.Identifier, uast.Expression).Children(
			On(HasInternalRole("dest")).Roles(uast.Call, uast.Argument, uast.Positional),
			On(HasInternalRole("nl")).Roles(uast.Call, uast.Argument, uast.Positional),
			On(HasInternalRole("values")).Roles(uast.Call, uast.Argument, uast.Positional).Children(
				On(Any).Roles(uast.Call, uast.Argument, uast.Positional),
			),
		),

		// Python annotations for variables, function argument or return values doesn't
		// have any semantic information by themselves and this we consider it comments
		// (some preprocessors or linters can use them, the runtimes ignore them). The
		// TOKEN will take the annotation in the UAST node so the information is keept in
		// any case.  FIXME: need annotation or type UAST roles
		On(pyast.AnnAssign).Roles(uast.Operator, uast.Binary, uast.Assignment, uast.Comment, uast.Incomplete),
		On(HasInternalRole("annotation")).Roles(uast.Comment, uast.Incomplete),
		On(HasInternalRole("returns")).Roles(uast.Comment, uast.Incomplete),

		// Python very odd ellipsis operatouast. Has a special rule in tonoder synthetic tokens
		// map to load it with the token "PythonEllipsisuast.Operator" and gets the role uast.Identifier
		On(pyast.Ellipsis).Roles(uast.Identifier, uast.Incomplete),

		// uast.List/uast.Map/uast.Set comprehensions. We map the "for x in y" to uast.For, uast.Iterator (foreach)
		// roles and the "if something" to uast.If* roles. FIXME: missing the top comprehension
		// roles in the UAST, change once they've been merged
		On(pyast.ListComp).Roles(uast.List, uast.For, uast.Expression),
		On(pyast.DictComp).Roles(uast.Map, uast.For, uast.Expression),
		On(pyast.SetComp).Roles(uast.Set, uast.For, uast.Expression),
		On(pyast.Comprehension).Roles(uast.For, uast.Iterator, uast.Expression, uast.Incomplete).Children(
			On(HasInternalRole("iter")).Roles(uast.For, uast.Update, uast.Statement),
			On(HasInternalRole("target")).Roles(uast.For, uast.Expression),
			// FIXME: see the comment on uast.If, uast.Condition above
			On(pyast.Compare).Roles(uast.If, uast.Condition),
		),

		On(pyast.Delete).Roles(uast.Statement, uast.Incomplete),
		On(pyast.Await).Roles(uast.Statement, uast.Incomplete),
		On(pyast.Global).Roles(uast.Statement, uast.Visibility, uast.World, uast.Incomplete),
		On(pyast.Nonlocal).Roles(uast.Statement, uast.Visibility, uast.Module, uast.Incomplete),

		On(pyast.Yield).Roles(uast.Return, uast.Statement, uast.Incomplete),
		On(pyast.YieldFrom).Roles(uast.Return, uast.Statement, uast.Incomplete),
		On(pyast.Yield).Roles(uast.Literal, uast.List, uast.Expression, uast.Incomplete),

		On(pyast.Subscript).Roles(uast.Expression, uast.Incomplete),
		On(pyast.Index).Roles(uast.Expression, uast.Incomplete),
		On(pyast.Slice).Roles(uast.Expression, uast.Incomplete),
		On(pyast.ExtSlice).Roles(uast.Expression, uast.Incomplete),
	))
