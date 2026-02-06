// Package plan provides structures for representing resolved deployment plans.
//
// A plan is the result of resolving all references in a blueprint against
// available resources and their versions. It contains concrete deployment
// configuration ready for execution by a deployment provider.
//
// Plans are machine-generated and use JSON as their serialization format.
//
// Example usage:
//
//	p, err := plan.Read("plan.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = p.Write("output.json")
//	if err != nil {
//		log.Fatal(err)
//	}
package plan
