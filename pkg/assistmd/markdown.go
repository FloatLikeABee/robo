// Package assistmd formats assistant chat replies as Markdown for platform-chat rendering.
package assistmd

import (
	"fmt"
	"strings"
)

// Title returns a bold heading line.
func Title(text string) string {
	return "**" + strings.TrimSpace(text) + "**"
}

// Empty returns an italic empty-state line.
func Empty(text string) string {
	return "_" + strings.TrimSpace(text) + "_"
}

// Success returns a checkmark success line.
func Success(action, detail string) string {
	action = strings.TrimSpace(action)
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return "✓ **" + action + "**"
	}
	return fmt.Sprintf("✓ **%s** — %s", action, detail)
}

// BulletList builds a markdown bullet list with optional intro paragraph.
func BulletList(intro string, items []string) string {
	if len(items) == 0 {
		if intro != "" {
			return intro
		}
		return ""
	}
	var b strings.Builder
	if intro = strings.TrimSpace(intro); intro != "" {
		b.WriteString(intro)
		b.WriteString("\n\n")
	}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// KVBlock renders labeled key-value lines as a bullet list.
func KVBlock(title string, pairs [][2]string) string {
	items := make([]string, 0, len(pairs))
	for _, p := range pairs {
		k := strings.TrimSpace(p[0])
		v := strings.TrimSpace(p[1])
		if k == "" || v == "" {
			continue
		}
		items = append(items, fmt.Sprintf("**%s:** %s", k, v))
	}
	return BulletList(title, items)
}

// NamedSlug formats a list row with name and slug/path in backticks.
func NamedSlug(name, slug string) string {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if slug != "" {
		return fmt.Sprintf("**%s** — `%s`", name, slug)
	}
	return "**" + name + "**"
}

// NamedID formats a list row with name and numeric id.
func NamedID(name string, id any) string {
	return fmt.Sprintf("**%s** — #%v", strings.TrimSpace(name), id)
}
