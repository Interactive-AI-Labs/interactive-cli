package inputs

var DefaultProjectAPIKeyColumns = []string{"id", "public_key", "secret", "note", "created_at"}

var AllProjectAPIKeyColumns = []string{
	"id",
	"public_key",
	"secret",
	"note",
	"status",
	"expires_at",
	"last_used_at",
	"created_at",
}

var DefaultRouterAPIKeyColumns = []string{"id", "name", "key", "limit", "created_at"}

var AllRouterAPIKeyColumns = []string{
	"id",
	"name",
	"description",
	"status",
	"key",
	"disabled",
	"limit",
	"remaining",
	"limit_reset",
	"expires_at",
	"last_used_at",
	"created_at",
	"updated_at",
	"project_id",
	"user_id",
}
