// Package state provides structures for tracking deployment state.
//
// State records what resources have been deployed, their runtime identifiers,
// and their current operational status. It is used for incremental deployments
// and resource lifecycle management.
//
// State files are machine-generated and use JSON as their serialization format.
//
// Example usage:
//
//	st, err := state.Read("state.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = st.Write("state.json")
//	if err != nil {
//		log.Fatal(err)
//	}
package state
