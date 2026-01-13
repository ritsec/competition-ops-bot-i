package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Credential holds the schema definition for the Credential entity.
type Credential struct {
	ent.Schema
}

// Fields of the Credential.
func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.Int("number").
			Comment("Team number").
			Unique(),
		field.String("compsole").
			Comment("Compsole password"),
		field.String("scorify").
			Comment("Scorify password"),
		field.String("authentik").
			Comment("Authentik password"),
	}
}

// Edges of the Credential.
func (Credential) Edges() []ent.Edge {
	return nil
}
