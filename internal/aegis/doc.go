// Package aegis defines the Crucible Aegis policy DSL.
//
// Aegis is the language Crucible affordances use to declare permissions and
// provisioning requests. Each grant targets a subsystem (for example, ".cap
// effective net_admin") and carries positional arguments, keyword arguments,
// and an optional where clause. The package owns the lexer, the parser, and
// the AST node types that represent a parsed grant. The grant model used by
// manifests and the codec hooks that decode grants from manifest input live
// in the manifest package, which calls [Parse] when it needs the AST.
//
// Parsing a grant source string:
//
//	parsed, err := aegis.Parse(".cap effective net_admin")
//	if err != nil {
//		log.Fatal(err)
//	}
package aegis
