// Package publish is the in-process rule for anonymous Morph page reads.
// HTTP middleware and morph-mcp both call it. A record is public only when
// it was explicitly published (non-empty slug).
package publish

import "strings"

// Public page kinds. A new published route is added here and registered as GET.
var pageKinds = map[string]struct{}{
	"big-notes": {},
	"timelines": {},
	"research":  {},
}

// PageRoute reports whether method and path are an explicitly published page.
// Only GET and HEAD of /api/tran/public/{kind}/{slug} match. kind is one of
// big-notes, timelines, or research. slug is one non-empty segment, not "." or "..".
func PageRoute(method, path string) bool {
	if method != "GET" && method != "HEAD" {
		return false
	}
	rest, ok := strings.CutPrefix(path, "/api/tran/public/")
	if !ok || rest == "" {
		return false
	}
	kind, slug, ok := strings.Cut(rest, "/")
	if !ok || kind == "" || slug == "" || strings.Contains(slug, "/") {
		return false
	}
	if _, known := pageKinds[kind]; !known {
		return false
	}
	if dotSegment(kind) || dotSegment(slug) {
		return false
	}
	return true
}

func dotSegment(segment string) bool {
	return segment == "." || segment == ".." || strings.Contains(segment, "..")
}

// Visible reports whether a stored record may be shown to a caller with no session.
// Empty and whitespace slugs are unpublished.
func Visible(publishedSlug string) bool {
	return strings.TrimSpace(publishedSlug) != ""
}
