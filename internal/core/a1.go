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
	for _, name := range []string{"person", "bicycle", "car", "phone", "water", "letter", "bird", "girl", "boy", "mother", "father", "door", "doctor", "child", "kitchen", "tree", "room", "mug", "rain"} {
		defs = append(defs, Definition{ID: ID("entity." + name), Kind: Entity})
	}
	for _, name := range []string{"have", "read", "see", "drink", "eat", "open", "close", "help", "write"} {
		role0 := "agent"
		role1 := "patient"
		if name == "have" {
			role0, role1 = "holder", "theme"
		}
		if name == "help" {
			role1 = "beneficiary"
		}
		defs = append(defs, Definition{ID: ID("predicate." + name), Kind: Predicate, Roles: []Role{{Name: role0, Target: Entity}, {Name: role1, Target: Entity}}})
	}
	for _, name := range []string{"sleep", "run", "fly", "sit", "work", "cook", "walk", "sing", "bark", "cry"} {
		defs = append(defs, Definition{ID: ID("predicate." + name), Kind: Predicate, Roles: []Role{{Name: "agent", Target: Entity}}})
	}
	defs = append(defs, Definition{ID: "relation.located", Kind: Predicate, StateRelation: true, Roles: []Role{{Name: "located", Target: Entity}, {Name: "site", Target: Entity}}})
	defs = append(defs, Definition{ID: "predicate.rain", Kind: Predicate, Roles: []Role{{Name: "theme", Target: Entity}}})
	// These spatial predicates describe different relations, not translated
	// preposition tokens. The site and located referents have exact roles.
	for _, relation := range []string{"inside", "on_support", "below", "at_site", "in_front", "behind", "adjacent"} {
		defs = append(defs, Definition{ID: ID("relation." + relation), Kind: Predicate, StateRelation: true,
			Roles: []Role{{Name: "located", Target: Entity}, {Name: "site", Target: Entity}}})
	}
	for _, name := range []string{"black", "small", "new", "big", "open", "red", "hot"} {
		defs = append(defs, Definition{ID: ID("property." + name), Kind: Property})
	}
	return NewCatalog(defs)
}
