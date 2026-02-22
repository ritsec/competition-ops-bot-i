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

// Fields of the Credential.
func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.String("compsole").
			Comment("Compsole password"),
		field.String("scorify").
			Comment("Scorify password"),
		field.String("authentik").
			Comment("Authentik password"),
		field.String("store").
			Comment("Store"),
		field.String("ctfd").
			Comment("CTFd password"),
		field.String("wazuh").
			Comment("Wazuh password"),
		field.String("pfsense").
			Comment("pfSense password"),
		field.String("default").
			Comment("Default password"),
		field.String("kali").
			Comment("Kali password"),
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
