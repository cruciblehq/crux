// Package blueprint provides structures for defining system compositions.
//
// A blueprint orchestrates how resources are deployed, declaring which
// services and widgets should be included in a system deployment. It
// serves as the input to the planning phase, where references are resolved
// and a concrete deployment plan is generated.
//
// Load a blueprint and generate a plan:
//
//	bp, err := blueprint.Read("blueprint.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	p, err := bp.Execute(ctx, blueprint.ExecuteOptions{
//		Registry: "http://hub:8080",
//		Provider: config.ProviderTypeLocal,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = p.Write("plan.json")
package blueprint
