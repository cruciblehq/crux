// Package provision implements the compute resource requirements subsystem.
//
// Provision grants declare how much CPU, memory, and disk a service needs.
// The blueprint builder reads the accumulated [Spec] after processing all of
// a service's affordances, then uses it to bin-pack services onto compute
// units and select concrete instance types.
//
// Grants have the form ".provision KEY=VALUE ..." where accepted keys are
// cpu, memory, and disk. CPU may be an integer vCPU count (e.g. 2) or a
// millicore quantity (e.g. 500m). Memory and disk are byte quantities with
// IEC binary (8Gi) or SI decimal (8G) suffixes. Multiple grants are additive.
//
//	s := provision.New(&provision.Spec{})
//	err := s.Build(parsed)
//	spec := s.Spec()
package provision
