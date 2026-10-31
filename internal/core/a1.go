package core

// BaseA1 is a language-neutral starter ontology. Its symbols are semantic
// definitions, never words or permitted translation pairs.
func BaseA1() (*Catalog, error) {
	defs := []Definition{
		{ID: "predicate.is_a", Kind: Predicate, ValueKind: Entity, Roles: []Role{{Name: "member", Target: Entity}}},
		{ID: "predicate.has_property", Kind: Predicate, ValueKind: Property, Roles: []Role{{Name: "bearer", Target: Entity}}},
		{ID: "entity.unspecified", Kind: Entity},
	}
	for _, name := range []string{"cat", "dog", "book", "house", "table", "chair", "teacher", "student", "apple", "window", "bag", "coffee", "school"} {
		defs = append(defs, Definition{ID: ID("entity." + name), Kind: Entity})
	}
	for _, name := range []string{"black", "small", "new", "big", "open", "red", "hot"} {
		defs = append(defs, Definition{ID: ID("property." + name), Kind: Property})
	}
	return NewCatalog(defs)
}
