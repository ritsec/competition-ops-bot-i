package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").
			Comment("Discord ID of user").
			Unique(),
		field.String("username").
			Comment("Discord username"),
		field.Bool("lead").
			Comment("If user is a lead of their team").
			Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("user").
			Unique(), // each user belongs to ONE team
	}
}
