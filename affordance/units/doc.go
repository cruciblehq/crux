// Package units provides unit suffix types and multipliers for quantity parsing.
//
// Quantity literals in AGL grants (e.g. "8Gi", "500m") carry a suffix that
// determines the multiplier applied to the integer part. This package defines
// the recognised suffix set and their integer scale factors so that subsystems
// can share the same parsing logic without duplicating tables.
//
// Determining the multiplier for a suffix:
//
//	mul, ok := units.QuantitySuffix("Gi").Multiplier()
package units
