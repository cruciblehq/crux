package store

import "embed"

//go:embed sql/*.sql sql/**/*.sql
var sqlFS embed.FS

// Reads a SQL file from the embedded filesystem. Panics if the file doesn't
// exist, catching missing files at init time rather than at runtime.
func mustReadSQL(path string) string {
	data, err := sqlFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}

var sqlSchema = mustReadSQL("sql/schema.sql")

var (
	sqlNamespacesDelete = mustReadSQL("sql/namespaces/delete.sql")
	sqlNamespacesGet    = mustReadSQL("sql/namespaces/get.sql")
	sqlNamespacesInsert = mustReadSQL("sql/namespaces/insert.sql")
)

var (
	sqlResourceSummariesInsert = mustReadSQL("sql/resource_summaries/insert.sql")
	sqlResourceSummariesList   = mustReadSQL("sql/resource_summaries/list.sql")
)

var (
	sqlResourcesDelete = mustReadSQL("sql/resources/delete.sql")
	sqlResourcesGet    = mustReadSQL("sql/resources/get.sql")
	sqlResourcesInsert = mustReadSQL("sql/resources/insert.sql")
)

var (
	sqlVersionsInsert = mustReadSQL("sql/versions/insert.sql")
	sqlVersionsList   = mustReadSQL("sql/versions/list.sql")
)

var (
	sqlChannelsInsert = mustReadSQL("sql/channels/insert.sql")
	sqlChannelsList   = mustReadSQL("sql/channels/list.sql")
)

var (
	sqlArchivesDelete = mustReadSQL("sql/archives/delete.sql")
	sqlArchivesGet    = mustReadSQL("sql/archives/get.sql")
	sqlArchivesInsert = mustReadSQL("sql/archives/insert.sql")
)
