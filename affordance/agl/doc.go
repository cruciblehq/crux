// Package agl implements the Affordance Grant Language (AGL) parser.
//
// AGL is the language Crucible affordances use to declare permissions and
// provisioning requests. Each grant targets a subsystem (for example, ".cap
// effective net_admin") and carries positional arguments, keyword arguments,
// and an optional where clause. The package owns the lexer, the parser, and
// the AST node types that represent a parsed grant.
//
// Parsing a grant source string:
//
//	parsed, err := agl.Parse(".cap effective net_admin")
//	if err != nil {
//		log.Fatal(err)
//	}
package agl
