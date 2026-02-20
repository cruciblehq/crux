// Package blueprint generates deployment plans from blueprints.
//
// Data structures and codec are defined in the spec module
// (github.com/cruciblehq/spec/blueprint). This package provides [Execute]
// to generate a deployment plan from a decoded blueprint.
//
//	data, err := os.ReadFile("blueprint.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bp, err := spec.Decode(data)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	p, err := blueprint.Execute(ctx, bp, blueprint.ExecuteOptions{
//		Registry: "http://hub:8080",
//		Provider: config.ProviderTypeLocal,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	out, _ := json.MarshalIndent(p, "", "  ")
//	os.WriteFile("plan.json", out, 0644)
package blueprint
