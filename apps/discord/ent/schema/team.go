package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Team holds the schema definition for the Team entity.
type Team struct {
	ent.Schema
}

// Fields of the Team.
func (Team) Fields() []ent.Field {
	return []ent.Field{
		field.String("lead").
			Comment("Discord username of team lead (if applicable)").
			Default("none"),
		field.Enum("type").
			Values(
				"blue",
				"red",
				"black",
				"white",
				"purple",
			).
			Comment("Type of team").
			Default("black"),
		field.Int("number").
			Comment("Team number").
			Optional(),
	}
}

// Edges of the Team.
func (Team) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type),
	}
}
