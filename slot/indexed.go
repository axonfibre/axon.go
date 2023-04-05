package slot

type IndexedID interface {
	comparable

	Index() Index
	String() string
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region IndexedEntity ////////////////////////////////////////////////////////////////////////////////////////////////

type IndexedEntity[IDType IndexedID] interface {
	ID() IDType
}
