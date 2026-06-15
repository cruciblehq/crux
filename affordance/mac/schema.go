package mac

// Classifies the value type of a hook field.
//
// The type determines which comparison operators and expression forms are
// valid. E.g., "like" requires TypeString and "between" requires TypeUint64.
type FieldType string

const (
	TypeUint64 FieldType = "uint64" // Number (uid, gid, ino, port).
	TypeString FieldType = "string" // String (path, comm, name).
	TypeBool   FieldType = "bool"   // Boolean flag.
)

// Single observable field exposed by an LSM hook.
//
// Fields are the leaf values that affordance rules can reference in where
// clauses. Some fields are only available in sleepable hooks because
// reading them may block (e.g. file.ima_hash requires disk I/O).
type Field struct {
	Name      string    // Dotted field name such as "file.path" or "task.uid".
	Type      FieldType // Value type used for validation and type checking.
	Sleepable bool      // True when this field is only available in sleepable hook contexts.
}

// Definition of a single LSM hook and the fields it exposes.
//
// Sleepable hooks run in process context and may perform blocking operations.
// Non-sleepable hooks run in atomic context and cannot access fields marked as
// sleepable.
type Hook struct {
	Name      string           // Kernel hook name such as "file_open" or "socket_bind".
	Sleepable bool             // Whether the hook runs in a sleepable (process) context.
	Fields    map[string]Field // Observable fields keyed by dotted name.
}

// Mapping from hook names to their definitions.
//
// The registry is the source of truth used by the validator to check
// that hook names, field references, and types are valid.
type Registry struct {
	hooks map[string]*Hook // Internal map of hook name to hook definition.
}

// Creates an empty registry with no hooks.
//
// Use AddHook to populate the registry, or call catalog to get a
// pre-populated registry with all known LSM hooks.
func NewRegistry() *Registry {
	return &Registry{hooks: make(map[string]*Hook)}
}

// Registers a hook definition in the registry.
//
// If a hook with the same name already exists it is silently replaced.
func (r *Registry) AddHook(h Hook) {
	r.hooks[h.Name] = &h
}

// Returns the hook definition for the given name, or nil if not found.
//
// Callers should check the return value before accessing hook fields.
func (r *Registry) LookupHook(name string) *Hook {
	return r.hooks[name]
}
