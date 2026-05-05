package aegis

import (
	"errors"
	"testing"
)

// Asserts that Parse succeeds and returns the resulting Model.
func parseOK(t *testing.T, src string) *Model {
	t.Helper()
	g, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): unexpected error: %v", src, err)
	}
	return g
}

// Asserts that Parse fails with an error wrapping ErrParse.
func parseErr(t *testing.T, src string) {
	t.Helper()
	_, err := Parse(src)
	if err == nil {
		t.Fatalf("Parse(%q): expected error, got nil", src)
	}
	if !errors.Is(err, ErrParse) {
		t.Fatalf("Parse(%q): error %v does not wrap ErrParse", src, err)
	}
}

// --- Grant header ---

func TestSimpleGrant(t *testing.T) {
	g := parseOK(t, ".seccomp read")
	if g.Subsystem != "seccomp" {
		t.Errorf("Subsystem: got %q, want %q", g.Subsystem, "seccomp")
	}
	if len(g.Args) != 1 {
		t.Fatalf("Args: got %d, want 1", len(g.Args))
	}
	if g.Args[0].Value != "read" || g.Args[0].Type != ArgName {
		t.Errorf("Args[0]: got %+v", g.Args[0])
	}
	if len(g.Kwargs) != 0 {
		t.Errorf("Kwargs: got %d, want 0", len(g.Kwargs))
	}
	if g.Where != nil {
		t.Errorf("Where: got non-nil, want nil")
	}
}

func TestGrantWithExtraName(t *testing.T) {
	g := parseOK(t, ".cap net_admin effective")
	if g.Subsystem != "cap" {
		t.Errorf("Subsystem: got %q, want %q", g.Subsystem, "cap")
	}
	if len(g.Args) != 2 {
		t.Fatalf("Args: got %d, want 2", len(g.Args))
	}
	if g.Args[0].Value != "net_admin" || g.Args[1].Value != "effective" {
		t.Errorf("Args: got %+v", g.Args)
	}
	if g.Args[1].Type != ArgName {
		t.Errorf("Args[1].Type: got %v, want ArgName", g.Args[1].Type)
	}
}

func TestGrantWithIntArgs(t *testing.T) {
	g := parseOK(t, ".rlimit nofile 1024 2048")
	if len(g.Args) != 3 {
		t.Fatalf("Args: got %d, want 3", len(g.Args))
	}
	if g.Args[0].Value != "nofile" || g.Args[0].Type != ArgName {
		t.Errorf("Args[0]: got %+v", g.Args[0])
	}
	if g.Args[1].Value != "1024" || g.Args[1].Type != ArgInt {
		t.Errorf("Args[1]: got %+v", g.Args[1])
	}
	if g.Args[2].Value != "2048" || g.Args[2].Type != ArgInt {
		t.Errorf("Args[2]: got %+v", g.Args[2])
	}
}

func TestGrantWithDottedName(t *testing.T) {
	g := parseOK(t, ".cgroup cpu.weight 100")
	if g.Args[0].Value != "cpu.weight" {
		t.Errorf("Args[0]: got %q, want %q", g.Args[0].Value, "cpu.weight")
	}
	if g.Args[1].Value != "100" || g.Args[1].Type != ArgInt {
		t.Errorf("Args[1]: got %+v", g.Args[1])
	}
}

// --- Kwargs ---

func TestGrantKwarg(t *testing.T) {
	g := parseOK(t, ".cgroup io.max 8 0 rbps=1048576")
	if len(g.Args) != 3 {
		t.Fatalf("Args: got %d, want 3: %v", len(g.Args), g.Args)
	}
	if g.Args[0].Value != "io.max" || g.Args[1].Value != "8" || g.Args[2].Value != "0" {
		t.Errorf("Args: got %+v", g.Args)
	}
	if len(g.Kwargs) != 1 {
		t.Fatalf("Kwargs: got %d, want 1", len(g.Kwargs))
	}
	kw := g.Kwargs[0]
	if kw.Key != "rbps" {
		t.Errorf("Kwargs[0].Key: got %q, want %q", kw.Key, "rbps")
	}
	if kw.Value.Type != ArgInt || kw.Value.Value != "1048576" {
		t.Errorf("Kwargs[0].Value: got %+v", kw.Value)
	}
}

func TestGrantMultipleKwargs(t *testing.T) {
	g := parseOK(t, `.net listen port=443 proto="tcp"`)
	if len(g.Args) != 1 {
		t.Fatalf("Args: got %d, want 1", len(g.Args))
	}
	if len(g.Kwargs) != 2 {
		t.Fatalf("Kwargs: got %d, want 2", len(g.Kwargs))
	}
	if g.Kwargs[0].Key != "port" || g.Kwargs[0].Value.Value != "443" {
		t.Errorf("Kwargs[0]: got %+v", g.Kwargs[0])
	}
	if g.Kwargs[1].Key != "proto" || g.Kwargs[1].Value.Value != "tcp" {
		t.Errorf("Kwargs[1]: got %+v", g.Kwargs[1])
	}
	if g.Kwargs[1].Value.Type != ArgStrASCII {
		t.Errorf("Kwargs[1].Value.Type: got %v, want ArgStrASCII", g.Kwargs[1].Value.Type)
	}
}

func TestGrantKwargWithName(t *testing.T) {
	g := parseOK(t, ".net connect host=example.com")
	if len(g.Kwargs) != 1 {
		t.Fatalf("Kwargs: got %d", len(g.Kwargs))
	}
	if g.Kwargs[0].Value.Type != ArgName || g.Kwargs[0].Value.Value != "example.com" {
		t.Errorf("Kwargs[0].Value: got %+v", g.Kwargs[0].Value)
	}
}

func TestErrorPositionalAfterKwarg(t *testing.T) {
	parseErr(t, ".net listen port=443 tcp")
}

func TestErrorKwargWithoutValue(t *testing.T) {
	parseErr(t, ".net listen port=")
}

// --- Quoted strings as positional args ---

func TestStringQuotedPositional(t *testing.T) {
	g := parseOK(t, ".cgroup cpuset.cpus \"0-3,5\"")
	if len(g.Args) != 2 {
		t.Fatalf("Args: got %d, want 2: %v", len(g.Args), g.Args)
	}
	if g.Args[1].Value != "0-3,5" || g.Args[1].Type != ArgStrASCII {
		t.Errorf("Args[1]: got %+v", g.Args[1])
	}
}

func TestStringPath(t *testing.T) {
	g := parseOK(t, ".fcap net_raw /usr/bin/ping effective")
	if len(g.Args) != 3 {
		t.Fatalf("Args: got %d, want 3: %v", len(g.Args), g.Args)
	}
	if g.Args[1].Value != "/usr/bin/ping" || g.Args[1].Type != ArgStrASCII {
		t.Errorf("Args[1]: got %+v", g.Args[1])
	}
	if g.Args[2].Value != "effective" || g.Args[2].Type != ArgName {
		t.Errorf("Args[2]: got %+v", g.Args[2])
	}
}

func TestStringQuotedWithSpaces(t *testing.T) {
	g := parseOK(t, `.fcap net_raw "/opt/my app/bin" effective`)
	if g.Args[1].Value != "/opt/my app/bin" {
		t.Errorf("Args[1]: got %q", g.Args[1].Value)
	}
}

func TestStringEscapes(t *testing.T) {
	g := parseOK(t, `.fcap net_raw "say \"hi\""`)
	if g.Args[1].Value != `say "hi"` {
		t.Errorf("Args[1]: got %q, want %q", g.Args[1].Value, `say "hi"`)
	}
}

func TestUnicodeStringQuoted(t *testing.T) {
	g := parseOK(t, ".fcap net_raw u\"/opt/café/my app\" effective")
	if g.Args[1].Value != "/opt/café/my app" {
		t.Errorf("Args[1]: got %q", g.Args[1].Value)
	}
	if g.Args[1].Type != ArgStrUnicode {
		t.Errorf("Args[1].Type: got %v, want ArgStrUnicode", g.Args[1].Type)
	}
}

// --- Expression: comparisons ---

func TestWhereFieldCompare(t *testing.T) {
	g := parseOK(t, ".mac socket_bind where socket.port >= 1024")
	cmp, ok := g.Where.(*CompareExpr)
	if !ok {
		t.Fatalf("Where: got %T", g.Where)
	}
	if !cmp.Left.IsField || cmp.Left.Field != "socket.port" {
		t.Errorf("Left: got %+v", cmp.Left)
	}
	if cmp.Op != CmpGte {
		t.Errorf("Op: got %s, want >=", cmp.Op)
	}
	if cmp.Right.Value.Type != ValueInt || cmp.Right.Value.Int != 1024 {
		t.Errorf("Right: got %+v", cmp.Right)
	}
}

func TestWhereFieldToField(t *testing.T) {
	g := parseOK(t, ".mac ptrace_access_check where task.uid = target.uid")
	cmp := g.Where.(*CompareExpr)
	if !cmp.Left.IsField || cmp.Left.Field != "task.uid" {
		t.Errorf("Left: got %+v", cmp.Left)
	}
	if !cmp.Right.IsField || cmp.Right.Field != "target.uid" {
		t.Errorf("Right: got %+v", cmp.Right)
	}
}

// --- Expression: like / not like ---

func TestWhereLike(t *testing.T) {
	g := parseOK(t, `.mac file_open where file.path like "/usr/bin/**"`)
	like := g.Where.(*LikeExpr)
	if like.Negated {
		t.Errorf("Negated: got true, want false")
	}
	if like.Pattern != "/usr/bin/**" {
		t.Errorf("Pattern: got %q", like.Pattern)
	}
}

func TestWhereNotLike(t *testing.T) {
	g := parseOK(t, `.mac file_open where file.path not like "/tmp/**"`)
	like := g.Where.(*LikeExpr)
	if !like.Negated {
		t.Errorf("Negated: got false, want true")
	}
	if like.Pattern != "/tmp/**" {
		t.Errorf("Pattern: got %q", like.Pattern)
	}
}

func TestWhereLikeWithEscape(t *testing.T) {
	g := parseOK(t, `.mac file_open where file.name like "with\\backslash"`)
	like := g.Where.(*LikeExpr)
	if like.Pattern != `with\backslash` {
		t.Errorf("Pattern: got %q, want %q", like.Pattern, `with\backslash`)
	}
}

func TestStringInLikeUnquoted(t *testing.T) {
	g := parseOK(t, ".mac file_open where file.path like /usr/bin/**")
	like := g.Where.(*LikeExpr)
	if like.Pattern != "/usr/bin/**" {
		t.Errorf("Pattern: got %q", like.Pattern)
	}
}

// --- Expression: in ---

func TestWhereIn(t *testing.T) {
	g := parseOK(t, ".seccomp ioctl where arg1 in (0x5401, 0x5402, 0x540e)")
	in := g.Where.(*InExpr)
	if !in.Field.IsField || in.Field.Field != "arg1" {
		t.Errorf("Field: got %+v", in.Field)
	}
	want := []uint64{0x5401, 0x5402, 0x540e}
	for i, v := range in.Values {
		if v.Value.Int != want[i] {
			t.Errorf("Values[%d]: got %#x, want %#x", i, v.Value.Int, want[i])
		}
	}
}

func TestWhereInStrings(t *testing.T) {
	g := parseOK(t, `.mac file_open where file.path in ("/etc/passwd", "/etc/shadow")`)
	in := g.Where.(*InExpr)
	if in.Values[0].Value.Str != "/etc/passwd" {
		t.Errorf("Values[0]: got %q", in.Values[0].Value.Str)
	}
	if in.Values[1].Value.Str != "/etc/shadow" {
		t.Errorf("Values[1]: got %q", in.Values[1].Value.Str)
	}
}

func TestWhereInPaths(t *testing.T) {
	g := parseOK(t, ".mac file_open where file.path in (/etc/passwd, /usr/bin/ping)")
	in := g.Where.(*InExpr)
	if in.Values[0].Value.Str != "/etc/passwd" {
		t.Errorf("Values[0]: got %q", in.Values[0].Value.Str)
	}
	if in.Values[1].Value.Str != "/usr/bin/ping" {
		t.Errorf("Values[1]: got %q", in.Values[1].Value.Str)
	}
}

// --- Expression: between ---

func TestWhereBetween(t *testing.T) {
	g := parseOK(t, ".mac socket_bind where socket.port between 1024 and 65535")
	between := g.Where.(*BetweenExpr)
	if between.Low.Value.Int != 1024 {
		t.Errorf("Low: got %d", between.Low.Value.Int)
	}
	if between.High.Value.Int != 65535 {
		t.Errorf("High: got %d", between.High.Value.Int)
	}
}

// --- Expression: bitwise ---

func TestWhereBitwiseTruthTest(t *testing.T) {
	g := parseOK(t, ".seccomp mmap where arg2 & 0x4")
	bit := g.Where.(*BitTestExpr)
	if bit.Val != nil {
		t.Errorf("Val: got non-nil")
	}
	if bit.Mask.Value.Int != 0x4 {
		t.Errorf("Mask: got %d", bit.Mask.Value.Int)
	}
}

func TestWhereBitwiseEqualityTest(t *testing.T) {
	g := parseOK(t, ".seccomp mmap where arg2 & 0x4 = 0")
	bit := g.Where.(*BitTestExpr)
	if bit.Val == nil {
		t.Fatal("Val: got nil")
	}
	if bit.Val.Value.Int != 0 {
		t.Errorf("Val: got %d", bit.Val.Value.Int)
	}
}

// --- Expression: boolean combinators ---

func TestWhereAnd(t *testing.T) {
	g := parseOK(t, ".seccomp socket where arg0 in (1, 2) and arg1 in (1, 2)")
	bin := g.Where.(*BinaryExpr)
	if bin.Op != OpAnd {
		t.Errorf("Op: got %s", bin.Op)
	}
}

func TestWhereOr(t *testing.T) {
	g := parseOK(t, `.mac file_open where file.path like "/etc/**" or file.path like "/usr/**"`)
	bin := g.Where.(*BinaryExpr)
	if bin.Op != OpOr {
		t.Errorf("Op: got %s", bin.Op)
	}
}

func TestWhereNot(t *testing.T) {
	g := parseOK(t, ".mac file_open where not file.mode & 0x2")
	unary := g.Where.(*UnaryExpr)
	if unary.Op != OpNot {
		t.Errorf("Op: got %v", unary.Op)
	}
}

func TestWherePrecedence(t *testing.T) {
	g := parseOK(t, ".mac file_open where file.mode = 1 or file.mode = 2 and file.mode = 3")
	bin := g.Where.(*BinaryExpr)
	if bin.Op != OpOr {
		t.Errorf("top-level: got %s", bin.Op)
	}
	if _, ok := bin.Right.(*BinaryExpr); !ok {
		t.Errorf("Right: got %T", bin.Right)
	}
}

func TestWhereParen(t *testing.T) {
	g := parseOK(t, ".mac file_open where (file.mode = 1 or file.mode = 2) and file.mode = 3")
	bin := g.Where.(*BinaryExpr)
	if bin.Op != OpAnd {
		t.Errorf("top-level: got %s", bin.Op)
	}
}

// --- Error cases ---

func TestErrorNoDot(t *testing.T) {
	parseErr(t, "seccomp read")
}

func TestErrorWhitespaceAfterDot(t *testing.T) {
	// The subsystem header must be lexed as a single token; whitespace
	// between '.' and the name is not permitted.
	_, err := Parse(". seccomp read")
	if err == nil {
		t.Fatal("expected error for whitespace between '.' and subsystem name")
	}
}

func TestErrorNoFirstArg(t *testing.T) {
	parseErr(t, ".seccomp")
}

func TestErrorNoFirstArgBeforeWhere(t *testing.T) {
	parseErr(t, ".seccomp where arg0 = 1")
}

func TestErrorFirstArgNotName(t *testing.T) {
	// First positional must be a NAME, not an integer.
	parseErr(t, ".rlimit 1024")
}

func TestErrorFirstArgNotString(t *testing.T) {
	// First positional must be a NAME, not a string literal.
	parseErr(t, `.fcap "net_raw"`)
}

func TestErrorFirstArgNotPath(t *testing.T) {
	// First positional must be a NAME, not an unquoted path.
	parseErr(t, ".fcap /usr/bin/ping")
}

func TestErrorPositionalDash(t *testing.T) {
	// '-' is not a valid positional token. Composite values like 0-3 must
	// be quoted to be a single string positional.
	parseErr(t, ".cgroup cpuset.cpus 0-3,5")
}

func TestErrorTrailingJunk(t *testing.T) {
	parseErr(t, ".seccomp read junk_after where")
}

func TestErrorBadEscape(t *testing.T) {
	_, err := Parse(`.mac file_open where file.path like "\q"`)
	if err == nil {
		t.Fatal("expected error for bad escape sequence")
	}
}

// --- Integer literal bases ---

func TestIntLiteralHex(t *testing.T) {
	g := parseOK(t, ".seccomp ioctl where arg1 in (0x5401, 0x5402)")
	in := g.Where.(*InExpr)
	if in.Values[0].Value.Int != 0x5401 || in.Values[1].Value.Int != 0x5402 {
		t.Errorf("Values: got %#x %#x", in.Values[0].Value.Int, in.Values[1].Value.Int)
	}
}

func TestIntLiteralOctalModern(t *testing.T) {
	g := parseOK(t, ".mac file_open where file.mode & 0o644")
	bit := g.Where.(*BitTestExpr)
	if bit.Mask.Value.Int != 0o644 {
		t.Errorf("Mask: got %#o", bit.Mask.Value.Int)
	}
}

func TestIntLiteralOctalLegacy(t *testing.T) {
	g := parseOK(t, ".mac file_open where file.mode & 0644")
	bit := g.Where.(*BitTestExpr)
	if bit.Mask.Value.Int != 0o644 {
		t.Errorf("Mask: got %#o", bit.Mask.Value.Int)
	}
}

func TestIntLiteralBinary(t *testing.T) {
	g := parseOK(t, ".seccomp mmap where arg2 & 0b10101010")
	bit := g.Where.(*BitTestExpr)
	if bit.Mask.Value.Int != 0b10101010 {
		t.Errorf("Mask: got %b", bit.Mask.Value.Int)
	}
}

func TestErrorInvalidOctalDigit(t *testing.T) {
	_, err := Parse(".seccomp mmap where arg2 & 09")
	if err == nil {
		t.Fatal("expected error for invalid octal digit '9'")
	}
}

// --- Lexer-level errors surfaced by Parse ---

func TestErrorAtIsInvalid(t *testing.T) {
	_, err := Parse(".fcap net_raw @/opt/bin")
	if err == nil {
		t.Fatal("expected error for '@'")
	}
}

func TestErrorPathWithNonASCII(t *testing.T) {
	_, err := Parse(".fcap net_raw /opt/café/bin")
	if err == nil {
		t.Fatal("expected error for non-ASCII in unquoted path")
	}
}

func TestErrorPathWithDisallowedChar(t *testing.T) {
	_, err := Parse(".mac file_open where file.path like /usr/bin/@tool")
	if err == nil {
		t.Fatal("expected error for '@' in unquoted path")
	}
}

func TestGlobBracketRejected(t *testing.T) {
	_, err := Parse(".mac file_open where file.path like /dev/sd[!a-z]")
	if err == nil {
		t.Fatal("expected error for '[' in unquoted glob")
	}
}

// --- Variables ---

func TestVarPositional(t *testing.T) {
	g := parseOK(t, ".cgroup cpu.max $cpumin $cpumax")
	if len(g.Args) != 3 {
		t.Fatalf("Args: got %d, want 3", len(g.Args))
	}
	if g.Args[1].Type != ArgVar || g.Args[1].Value != "cpumin" {
		t.Errorf("Args[1]: got %+v, want ArgVar \"cpumin\"", g.Args[1])
	}
	if g.Args[2].Type != ArgVar || g.Args[2].Value != "cpumax" {
		t.Errorf("Args[2]: got %+v, want ArgVar \"cpumax\"", g.Args[2])
	}
}

func TestVarKwargValue(t *testing.T) {
	g := parseOK(t, ".cgroup io.max 8 0 rbps=$limit")
	if len(g.Kwargs) != 1 {
		t.Fatalf("Kwargs: got %d, want 1", len(g.Kwargs))
	}
	kw := g.Kwargs[0]
	if kw.Key != "rbps" {
		t.Errorf("Kwargs[0].Key: got %q, want %q", kw.Key, "rbps")
	}
	if kw.Value.Type != ArgVar || kw.Value.Value != "limit" {
		t.Errorf("Kwargs[0].Value: got %+v, want ArgVar \"limit\"", kw.Value)
	}
}

func TestVarInWhere(t *testing.T) {
	g := parseOK(t, ".mac file_open where uid = $myuid")
	cmp, ok := g.Where.(*CompareExpr)
	if !ok {
		t.Fatalf("Where: got %T, want *CompareExpr", g.Where)
	}
	if cmp.Right.IsField {
		t.Fatalf("Right.IsField: got true, want false")
	}
	if cmp.Right.Value.Type != ValueVar || cmp.Right.Value.Str != "myuid" {
		t.Errorf("Right.Value: got %+v, want ValueVar \"myuid\"", cmp.Right.Value)
	}
}

func TestVarInInList(t *testing.T) {
	g := parseOK(t, ".mac file_open where uid in ($a, $b, 0)")
	in, ok := g.Where.(*InExpr)
	if !ok {
		t.Fatalf("Where: got %T, want *InExpr", g.Where)
	}
	if len(in.Values) != 3 {
		t.Fatalf("Values: got %d, want 3", len(in.Values))
	}
	if in.Values[0].Value.Type != ValueVar || in.Values[0].Value.Str != "a" {
		t.Errorf("Values[0]: got %+v", in.Values[0])
	}
	if in.Values[1].Value.Type != ValueVar || in.Values[1].Value.Str != "b" {
		t.Errorf("Values[1]: got %+v", in.Values[1])
	}
	if in.Values[2].Value.Type != ValueInt || in.Values[2].Value.Int != 0 {
		t.Errorf("Values[2]: got %+v", in.Values[2])
	}
}

func TestVarRejectedAsSubsystem(t *testing.T) {
	// The lexer rejects '.' followed by anything other than a name-start, so
	// ".$sub" is a lex error rather than a parse error.
	_, err := Parse(".$sub read")
	if err == nil {
		t.Fatal("expected error for '$' as subsystem name")
	}
}

func TestVarLexErrorPropagates(t *testing.T) {
	_, err := Parse(".cgroup cpu.max $")
	if err == nil {
		t.Fatal("expected error for bare '$'")
	}
}
