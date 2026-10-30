package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Credential holds the schema definition for the Credential entity.
type Credential struct {
	ent.Schema
}

// ServiceCredential is a team's login for a single competition service
type ServiceCredential struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Fields of the Credential.
func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.String("password").
			Sensitive().
			Comment("Team password"),
		field.JSON("services", []ServiceCredential{}).
			Comment("Logins for each competition service, with team values filled in"),
	}
}

// Edges of the Credential.
func (Credential) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("credential").
			Unique(),
	}
}
