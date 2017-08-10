package normalizer

import (
	"errors"

	"github.com/bblfsh/python-driver/driver/normalizer/pyast"

	. "github.com/bblfsh/sdk/uast"
	. "github.com/bblfsh/sdk/uast/ann"
)

/*
   Tip: to quickly see the native AST generated by Python you can do:
   from ast import *
   code = 'print("test code")'
   dump(parse(code))
*/

/*
For a description of Python AST nodes:

https://greentreesnakes.readthedocs.io/en/latest/nodes.html

*/

/*
Unmarked nodes or nodes needing new features from the SDK:

   These nodes would need a list-mix feature to convert parallel lists
   into list of parents and children:

   BoolOp
   arguments.defaults
   arguments.keywords: same
   Compare.comparators
   Compare.ops
   IfCondition.left
	(see: https://greentreesnakes.readthedocs.io/en/latest/nodes.html#Compare)
*/

var AnnotationRules = On(Any).Self(
	On(Not(HasInternalType(pyast.Module))).Error(errors.New("root must be Module")),
	On(HasInternalType(pyast.Module)).Roles(File).Descendants(
		// Binary Expressions
		On(HasInternalType(pyast.BinOp)).Roles(BinaryExpression, Expression).Children(
			On(HasInternalRole("op")).Roles(BinaryExpressionOp),
			On(HasInternalRole("left")).Roles(BinaryExpressionLeft),
			On(HasInternalRole("right")).Roles(BinaryExpressionRight),
		),
		// Comparison operators
		On(HasInternalType(pyast.Eq)).Roles(OpEqual),
		On(HasInternalType(pyast.NotEq)).Roles(OpNotEqual),
		On(HasInternalType(pyast.Lt)).Roles(OpLessThan),
		On(HasInternalType(pyast.LtE)).Roles(OpLessThanEqual),
		On(HasInternalType(pyast.Gt)).Roles(OpGreaterThan),
		On(HasInternalType(pyast.GtE)).Roles(OpGreaterThanEqual),
		On(HasInternalType(pyast.Is)).Roles(OpSame),
		On(HasInternalType(pyast.IsNot)).Roles(OpNotSame),
		On(HasInternalType(pyast.In)).Roles(OpContains),
		On(HasInternalType(pyast.NotIn)).Roles(OpNotContains),

		// Aritmetic operators
		On(HasInternalType(pyast.Add)).Roles(OpAdd),
		On(HasInternalType(pyast.Sub)).Roles(OpSubstract),
		On(HasInternalType(pyast.Mult)).Roles(OpMultiply),
		On(HasInternalType(pyast.Div)).Roles(OpDivide),
		On(HasInternalType(pyast.Mod)).Roles(OpMod),
		On(HasInternalType(pyast.FloorDiv)).Roles(OpDivide, Incomplete),
		On(HasInternalType(pyast.Pow)).Roles(Incomplete),
		On(HasInternalType(pyast.MatMult)).Roles(OpMultiply, Incomplete),

		// Bitwise operators
		On(HasInternalType(pyast.LShift)).Roles(OpBitwiseLeftShift),
		On(HasInternalType(pyast.RShift)).Roles(OpBitwiseRightShift),
		On(HasInternalType(pyast.BitOr)).Roles(OpBitwiseOr),
		On(HasInternalType(pyast.BitXor)).Roles(OpBitwiseXor),
		On(HasInternalType(pyast.BitAnd)).Roles(OpBitwiseAnd),

		// Boolean operators
		On(HasInternalType(pyast.And)).Roles(OpBooleanAnd),
		On(HasInternalType(pyast.Or)).Roles(OpBooleanOr),
		On(HasInternalType(pyast.Not)).Roles(OpBooleanNot),

		On(HasInternalType(pyast.UnaryOp)).Roles(Expression, Incomplete),

		// Unary operators
		On(HasInternalType(pyast.Invert)).Roles(OpBitwiseComplement),
		On(HasInternalType(pyast.UAdd)).Roles(OpPositive),
		On(HasInternalType(pyast.USub)).Roles(OpNegative),

		On(HasInternalType(pyast.StringLiteral)).Roles(StringLiteral, Expression),
		On(HasInternalType(pyast.ByteLiteral)).Roles(ByteStringLiteral, Expression),
		On(HasInternalType(pyast.NumLiteral)).Roles(NumberLiteral, Expression),
		On(HasInternalType(pyast.Str)).Roles(StringLiteral, Expression),
		On(HasInternalType(pyast.BoolLiteral)).Roles(BooleanLiteral, Expression),
		On(HasInternalType(pyast.JoinedStr)).Roles(StringLiteral, Expression).Children(
			On(HasInternalType(pyast.FormattedValue)).Roles(Expression, Incomplete),
		),
		On(HasInternalType(pyast.NoneLiteral)).Roles(NullLiteral, Expression),
		On(HasInternalType(pyast.Set)).Roles(SetLiteral, Expression),
		On(HasInternalType(pyast.List)).Roles(ListLiteral, Expression),
		On(HasInternalType(pyast.Dict)).Roles(MapLiteral, Expression).Children(
			On(HasInternalRole("keys")).Roles(MapKey),
			On(HasInternalRole("values")).Roles(MapValue),
		),
		On(HasInternalType(pyast.Tuple)).Roles(TupleLiteral, Expression),

		// FIXME: the FunctionDeclarationReceiver is not set for methods; it should be taken from the parent
		// Type node Token (2 levels up) but the SDK doesn't allow this
		// TODO: create an issue for the SDK
		On(HasInternalType(pyast.FunctionDef)).Roles(FunctionDeclaration, FunctionDeclarationName,
			SimpleIdentifier),
		On(HasInternalType(pyast.AsyncFunctionDef)).Roles(FunctionDeclaration,
			FunctionDeclarationName, SimpleIdentifier, Incomplete),
		On(HasInternalType("FunctionDef.decorator_list")).Roles(Call, Incomplete),
		On(HasInternalType("FunctionDef.body")).Roles(FunctionDeclarationBody),
		// FIXME: change to FunctionDeclarationArgumentS once the PR has been merged
		On(HasInternalType(pyast.Arguments)).Roles(FunctionDeclarationArgument, Incomplete).Children(
			On(HasInternalRole("args")).Roles(FunctionDeclarationArgument, FunctionDeclarationArgumentName,
				SimpleIdentifier),
			On(HasInternalRole("vararg")).Roles(FunctionDeclarationArgument, FunctionDeclarationVarArgsList,
				FunctionDeclarationArgumentName, SimpleIdentifier),
			// FIXME: this is really different from vararg, change it when we have FunctionDeclarationMap
			// or something similar in the UAST
			On(HasInternalRole("kwarg")).Roles(FunctionDeclarationArgument, FunctionDeclarationVarArgsList,
				FunctionDeclarationArgumentName, Incomplete, SimpleIdentifier),
			// Default arguments: Python's AST puts default arguments on a sibling list to the one of
			// arguments that must be mapped to the arguments right-aligned like:
			// a, b=2, c=3 ->
			//		args    [a,b,c],
			//		defaults  [2,3]
			// TODO: create an issue for the SDK
			On(HasInternalType("arguments.defaults")).Roles(FunctionDeclarationArgumentDefaultValue, Incomplete),
			On(HasInternalType("arguments.keywords")).Roles(FunctionDeclarationArgumentDefaultValue, Incomplete),
			On(HasInternalType("AsyncFunctionDef.decorator_list")).Roles(Call, Incomplete),
			On(HasInternalType("AsyncFunctionDef.body")).Roles(FunctionDeclarationBody),
			// FIXME: change to FunctionDeclarationArgumentS once the PR has been merged
		),
		On(HasInternalType(pyast.Lambda)).Roles(FunctionDeclaration, Expression, Incomplete).Children(
			On(HasInternalType("Lambda.body")).Roles(FunctionDeclarationBody),
			// FIXME: change to FunctionDeclarationArgumentS once the PR has been merged
			On(HasInternalType(pyast.Arguments)).Roles(FunctionDeclarationArgument, Incomplete).Children(
				On(HasInternalRole("args")).Roles(FunctionDeclarationArgument, FunctionDeclarationArgumentName,
					SimpleIdentifier),
				On(HasInternalRole("vararg")).Roles(FunctionDeclarationArgument, FunctionDeclarationVarArgsList,
					FunctionDeclarationArgumentName, SimpleIdentifier),
				// FIXME: this is really different from vararg, change it when we have FunctionDeclarationMap
				// or something similar in the UAST
				On(HasInternalRole("kwarg")).Roles(FunctionDeclarationArgument, FunctionDeclarationVarArgsList,
					FunctionDeclarationArgumentName, Incomplete, SimpleIdentifier),
				// Default arguments: Python's AST puts default arguments on a sibling list to the one of
				// arguments that must be mapped to the arguments right-aligned like:
				// a, b=2, c=3 ->
				//		args    [a,b,c],
				//		defaults  [2,3]
				// TODO: create an issue for the SDK
				On(HasInternalType("arguments.defaults")).Roles(FunctionDeclarationArgumentDefaultValue,
					Incomplete),
				On(HasInternalType("arguments.keywords")).Roles(FunctionDeclarationArgumentDefaultValue,
					Incomplete),
			),
		),

		On(HasInternalType(pyast.Attribute)).Roles(SimpleIdentifier, Expression).Children(
			On(HasInternalType(pyast.Name)).Roles(QualifiedIdentifier)),

		On(HasInternalType(pyast.Call)).Roles(Call, Expression).Children(
			On(HasInternalRole("args")).Roles(CallPositionalArgument),
			On(HasInternalRole("keywords")).Roles(CallNamedArgument).Children(
				On(HasInternalRole("value")).Roles(CallNamedArgumentValue),
			),
			On(HasInternalRole("func")).Self(
				On(HasInternalType(pyast.Name)).Roles(Call),
				On(HasInternalType(pyast.Attribute)).Roles(CallCallee).Children(
					On(HasInternalRole("value")).Roles(CallReceiver),
				)),
		),

		//
		//	Assign => Assigment:
		//		targets[] => AssignmentVariable
		//		value	  => AssignmentValue
		//
		On(HasInternalType(pyast.Assign)).Roles(Assignment, Expression).Children(
			On(HasInternalRole("targets")).Roles(AssignmentVariable),
			On(HasInternalRole("value")).Roles(AssignmentValue),
		),

		On(HasInternalType(pyast.AugAssign)).Roles(AugmentedAssignment, Statement).Children(
			On(HasInternalRole("op")).Roles(AugmentedAssignmentOperator),
			On(HasInternalRole("target")).Roles(AugmentedAssignmentVariable),
			On(HasInternalRole("value")).Roles(AugmentedAssignmentValue),
		),

		On(HasInternalType(pyast.Expression)).Roles(Expression),
		On(HasInternalType(pyast.Expr)).Roles(Expression),
		On(HasInternalType(pyast.Name)).Roles(SimpleIdentifier, Expression),
		// Comments and non significative whitespace
		On(HasInternalType(pyast.SameLineNoops)).Roles(Comment),
		On(HasInternalType(pyast.PreviousNoops)).Roles(Whitespace).Children(
			On(HasInternalRole("lines")).Roles(Comment),
		),
		On(HasInternalType(pyast.RemainderNoops)).Roles(Whitespace).Children(
			On(HasInternalRole("lines")).Roles(Comment),
		),

		// TODO: check what Constant nodes are generated in the python AST and improve this
		On(HasInternalType(pyast.Constant)).Roles(SimpleIdentifier, Expression),
		On(HasInternalType(pyast.Try)).Roles(Try, Statement).Children(
			On(HasInternalRole("body")).Roles(TryBody),
			On(HasInternalRole("finalbody")).Roles(TryFinally),
			On(HasInternalRole("handlers")).Roles(TryCatch),
			On(HasInternalRole("orelse")).Roles(IfElse),
		),
		On(HasInternalType(pyast.TryExcept)).Roles(TryCatch, Statement),     // py2
		On(HasInternalType(pyast.ExceptHandler)).Roles(TryCatch, Statement), // py3
		On(HasInternalType(pyast.TryFinally)).Roles(TryFinally, Statement),
		On(HasInternalType(pyast.Raise)).Roles(Throw, Statement),
		// FIXME: review, add path for the body and items childs
		On(HasInternalType(pyast.With)).Roles(BlockScope, Statement),
		On(HasInternalType(pyast.AsyncWith)).Roles(BlockScope, Statement, Incomplete),
		On(HasInternalType(pyast.Withitem)).Roles(SimpleIdentifier, Incomplete),
		On(HasInternalType(pyast.Return)).Roles(Return, Statement),
		On(HasInternalType(pyast.Break)).Roles(Break, Statement),
		On(HasInternalType(pyast.Continue)).Roles(Continue, Statement),
		// FIXME: IfCondition bodies in Python take the form:
		// 1 < a < 10
		// - left (internalRole): 1 (first element)
		// - Compare.ops (internalType): [LessThan, LessThan]
		// - Compare.comparators (internalType): ['a', 10]
		// The current mapping is:
		// - left: BinaryExpressionLeft
		// - Compare.ops: BinaryExpressionOp
		// - Compare.comparators: BinaryExpressionRight
		// But this is obviously not correct. To fix this properly we would need
		// and SDK feature to mix lists (also needed for default and keyword arguments and
		// boolean operators).
		// "If that sounds awkward is because it is" (their words)
		On(HasInternalType(pyast.If)).Roles(If, Statement).Children(
			On(HasInternalType("If.body")).Roles(IfBody),
			On(HasInternalRole("test")).Roles(IfCondition),
			On(HasInternalType("If.orelse")).Roles(IfElse),
			On(HasInternalType(pyast.Compare)).Roles(BinaryExpression, Expression).Children(
				On(HasInternalType("Compare.ops")).Roles(BinaryExpressionOp),
				On(HasInternalType("Compare.comparators")).Roles(BinaryExpressionRight),
				On(HasInternalRole("left")).Roles(BinaryExpressionLeft),
			),
		),
		On(HasInternalType(pyast.IfExp)).Roles(If, Expression).Children(
			// These are used on ifexpressions (a = 1 if x else 2)
			On(HasInternalRole("body")).Roles(IfBody),
			On(HasInternalRole("test")).Roles(IfCondition),
			On(HasInternalRole("orelse")).Roles(IfElse),
		),
		On(HasInternalType(pyast.Import)).Roles(ImportDeclaration, Statement),
		// "y" in "from x import y" or "import y"
		On(HasInternalType(pyast.Alias)).Roles(ImportPath, SimpleIdentifier),
		// "x" in "from x import y"
		On(HasInternalType("ImportFrom.module")).Roles(ImportPath, SimpleIdentifier),
		// "y" in "import x as y"
		On(HasInternalType("alias.asname")).Roles(ImportAlias, SimpleIdentifier),
		On(HasInternalType(pyast.ImportFrom)).Roles(ImportDeclaration, Statement),
		On(HasInternalType(pyast.ClassDef)).Roles(TypeDeclaration, SimpleIdentifier, Statement).Children(
			On(HasInternalType("ClassDef.body")).Roles(TypeDeclarationBody),
			On(HasInternalType("ClassDef.bases")).Roles(TypeDeclarationBases),
		),

		On(HasInternalType(pyast.For)).Roles(ForEach, Statement).Children(
			On(HasInternalType("For.body")).Roles(ForBody),
			On(HasInternalRole("iter")).Roles(ForExpression),
			On(HasInternalRole("target")).Roles(ForUpdate),
			On(HasInternalType("For.orelse")).Roles(IfElse),
		),
		On(HasInternalType(pyast.AsyncFor)).Roles(ForEach, Statement, Incomplete).Children(
			On(HasInternalType("AsyncFor.body")).Roles(ForBody),
			On(HasInternalRole("iter")).Roles(ForExpression),
			On(HasInternalRole("target")).Roles(ForUpdate),
			On(HasInternalType("AsyncFor.orelse")).Roles(IfElse),
		),
		On(HasInternalType(pyast.While)).Roles(While, Statement).Children(
			On(HasInternalType("While.body")).Roles(WhileBody),
			On(HasInternalRole("test")).Roles(WhileCondition),
			On(HasInternalType("While.orelse")).Roles(IfElse),
		),
		On(HasInternalType(pyast.Pass)).Roles(Noop, Statement),
		On(HasInternalType(pyast.Num)).Roles(NumberLiteral, Expression),
		// FIXME: this is the annotated assignment (a: annotation = 3) not exactly Assignment
		// it also lacks AssignmentValue and AssignmentVariable (see how to add them)
		On(HasInternalType(pyast.Assert)).Roles(Assert, Statement),

		// These are AST nodes in Python2 but we convert them to functions in the UAST like
		// they are in Python3
		On(HasInternalType(pyast.Exec)).Roles(Call, Expression).Children(
			On(HasInternalRole("body")).Roles(CallPositionalArgument),
			On(HasInternalRole("globals")).Roles(CallPositionalArgument),
			On(HasInternalRole("locals")).Roles(CallPositionalArgument),
		),
		// Repr already comes as a Call \o/
		// Print as a function too.
		On(HasInternalType(pyast.Print)).Roles(Call, CallCallee, SimpleIdentifier, Expression).Children(
			On(HasInternalRole("dest")).Roles(CallPositionalArgument),
			On(HasInternalRole("nl")).Roles(CallPositionalArgument),
			On(HasInternalRole("values")).Roles(CallPositionalArgument).Children(
				On(Any).Roles(CallPositionalArgument),
			),
		),

		// Python annotations for variables, function argument or return values doesn't have any semantic
		// information by themselves and this we consider it comments (some preprocessors or linters can use
		// them, the runtimes ignore them). The TOKEN will take the annotation in the UAST node so
		// the information is keept in any case.
		// FIXME: need annotation or type UAST roles
		On(HasInternalType(pyast.AnnAssign)).Roles(Assignment, Comment, Incomplete),
		On(HasInternalType(pyast.Annotation)).Roles(Comment, Incomplete),
		On(HasInternalRole(pyast.Returns)).Roles(Comment, Incomplete),

		// Python very odd ellipsis operator. Has a special rule in tonoder synthetic tokens
		// map to load it with the token "PythonEllipsisOperator" and gets the role SimpleIdentifier
		On(HasInternalType(pyast.Ellipsis)).Roles(SimpleIdentifier, Incomplete),

		// List/Map/Set comprehensions. We map the "for x in y" to ForEach roles and the
		// "if something" to If* roles. FIXME: missing the top comprehension roles in the UAST, change
		// once they've been merged
		On(HasInternalType(pyast.Comprehension)).Roles(ForEach, Expression).Children(
			On(HasInternalRole("iter")).Roles(ForUpdate, Statement),
			On(HasInternalRole("target")).Roles(ForExpression),
			// FIXME: see the comment on IfCondition above
			On(HasInternalType(pyast.Compare)).Roles(IfCondition, BinaryExpression).Children(
				On(HasInternalType("Compare.ops")).Roles(BinaryExpressionOp),
				On(HasInternalType("Compare.comparators")).Roles(BinaryExpressionRight),
				On(HasInternalRole("left")).Roles(BinaryExpressionLeft),
			),
		),
		On(HasInternalType(pyast.ListComp)).Roles(ListLiteral, Expression, Incomplete),
		On(HasInternalType(pyast.SetComp)).Roles(MapLiteral, Expression, Incomplete),
		On(HasInternalType(pyast.SetComp)).Roles(SetLiteral, Expression, Incomplete),

		On(HasInternalType(pyast.Delete)).Roles(Statement, Incomplete),
		On(HasInternalType(pyast.Await)).Roles(Statement, Incomplete),
		On(HasInternalType(pyast.Global)).Roles(Statement, VisibleFromWorld, Incomplete),
		On(HasInternalType(pyast.Nonlocal)).Roles(Statement, VisibleFromModule, Incomplete),

		On(HasInternalType(pyast.Yield)).Roles(Return, Statement, Incomplete),
		On(HasInternalType(pyast.YieldFrom)).Roles(Return, Statement, Incomplete),
		On(HasInternalType(pyast.Yield)).Roles(ListLiteral, Expression, Incomplete),

		On(HasInternalType(pyast.Subscript)).Roles(Expression, Incomplete),
		On(HasInternalType(pyast.Index)).Roles(Expression, Incomplete),
		On(HasInternalType(pyast.Slice)).Roles(Expression, Incomplete),
		On(HasInternalType(pyast.ExtSlice)).Roles(Expression, Incomplete),
	),
)
