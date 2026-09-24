package mcp

import "idongivaflyinfa/publish"

// ExposeRecord is the in-process gate for a Morph record.
// A caller with no verified user id may receive the record only when it is
// published. A verified user keeps access to private records.
func ExposeRecord(caller Identity, publishedSlug string) bool {
	if caller.UserID == "" {
		return publish.Visible(publishedSlug)
	}
	return true
}
