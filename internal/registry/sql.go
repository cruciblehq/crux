package registry

import "embed"

//go:embed sql/*.sql sql/**/*.sql
var sqlFS embed.FS

// Reads a SQL file from the embedded filesystem, panicking if it doesn't exist.
func mustReadSQL(path string) string {
	data, err := sqlFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Schema definition for creating all registry tables.
//
// All foreign key constraints use ON DELETE RESTRICT to prevent accidental
// data loss. Deletion must be done bottom-up (channels first, then versions,
// then resources, then namespaces).
//
// Archive data is stored within the versions table as nullable columns (digest,
// size, path), populated when an archive is uploaded via UploadArchive.
var sqlSchema = mustReadSQL("sql/schema.sql")

var (
	sqlNamespacesInsert = mustReadSQL("sql/namespaces/insert.sql") // Insert new namespace
	sqlNamespacesGet    = mustReadSQL("sql/namespaces/get.sql")    // Get namespace details
	sqlNamespacesList   = mustReadSQL("sql/namespaces/list.sql")   // List all namespaces
	sqlNamespacesUpdate = mustReadSQL("sql/namespaces/update.sql") // Update namespace description
	sqlNamespacesDelete = mustReadSQL("sql/namespaces/delete.sql") // Delete namespace (requires no resources)
)

var (
	sqlResourcesInsert = mustReadSQL("sql/resources/insert.sql") // Insert new resource
	sqlResourcesGet    = mustReadSQL("sql/resources/get.sql")    // Get resource details
	sqlResourcesList   = mustReadSQL("sql/resources/list.sql")   // List resources in namespace
	sqlResourcesUpdate = mustReadSQL("sql/resources/update.sql") // Update resource metadata
	sqlResourcesDelete = mustReadSQL("sql/resources/delete.sql") // Delete resource (requires no versions)
)

var (
	sqlVersionsInsert = mustReadSQL("sql/versions/insert.sql") // Insert new version (archive fields NULL)
	sqlVersionsGet    = mustReadSQL("sql/versions/get.sql")    // Get version with archive details
	sqlVersionsList   = mustReadSQL("sql/versions/list.sql")   // List versions for resource
	sqlVersionsUpdate = mustReadSQL("sql/versions/update.sql") // Update version metadata
	sqlVersionsUpload = mustReadSQL("sql/versions/upload.sql") // Update version with archive metadata
	sqlVersionsDelete = mustReadSQL("sql/versions/delete.sql") // Delete version (requires no channels)
)

var (
	sqlChannelsInsert = mustReadSQL("sql/channels/insert.sql") // Insert new channel
	sqlChannelsGet    = mustReadSQL("sql/channels/get.sql")    // Get channel with version details
	sqlChannelsList   = mustReadSQL("sql/channels/list.sql")   // List channels for resource
	sqlChannelsUpdate = mustReadSQL("sql/channels/update.sql") // Update channel metadata
	sqlChannelsDelete = mustReadSQL("sql/channels/delete.sql") // Delete channel
)
