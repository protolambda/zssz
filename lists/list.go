package lists

// For slices to be a valid SSZ list, they need a defined limit.
// Lists without such definition will still be able to be serialized and deserialized,
// but are not supported to be merkleized with HashTreeRoot.
type List interface {
	Limit() uint32
}
