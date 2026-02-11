package manifest

// Implemented by resource configs that need structural validation after decode.
type validator interface {
	validate() error
}
