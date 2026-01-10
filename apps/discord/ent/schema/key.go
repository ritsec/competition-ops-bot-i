package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Key holds the schema definition for the Key entity.
type Key struct {
	ent.Schema
}

// Fields of the Key.
func (Key) Fields() []ent.Field {
	return []ent.Field{
		field.Strings("keys").
			Comment("User's SSH public key(s)"),
	}
}

// Edges of the Key.
func (Key) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("key"),
	}
}
