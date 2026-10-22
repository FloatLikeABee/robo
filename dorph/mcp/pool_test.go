package mcp_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/mcp"
)

func TestShareConsentScannerPoolAndReport(t *testing.T) {
	mcp.ResetPoolForTest()
	t.Cleanup(mcp.ResetPoolForTest)
	path, _ := seedTwoUsers(t)
	ctx := context.Background()
	db, err := mcp.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := mcp.ShareNote(ctx, db, "ada-id", ""); !errors.Is(err, mcp.ErrConsentRequired) {
		t.Fatalf("empty token = %v", err)
	}
	if n := shareCount(t, db); n != 0 {
		t.Fatalf("consent wrote %d shares", n)
	}

	clean, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "Dock brief", Body: "north gate is open at dawn"})
	if err != nil {
		t.Fatal(err)
	}
	prev, err := mcp.PreviewShare(ctx, db, "ada-id", clean.RecordID, "pool", "", []string{"productivity"}, []string{"dock"})
	if err != nil || prev.ConfirmToken == "" || prev.Alias == "ada@example.com" || strings.Contains(prev.Content, "ada@example.com") {
		t.Fatalf("preview = %+v err=%v", prev, err)
	}
	if _, err := mcp.ShareNote(ctx, db, "bea-id", prev.ConfirmToken); !errors.Is(err, mcp.ErrConsentRequired) {
		t.Fatalf("foreign token = %v", err)
	}

	prev, err = mcp.PreviewShare(ctx, db, "ada-id", clean.RecordID, "pool", "", []string{"productivity"}, []string{"dock"})
	if err != nil {
		t.Fatal(err)
	}
	mcp.ExpireConfirmTokenForTest(prev.ConfirmToken)
	if _, err := mcp.ShareNote(ctx, db, "ada-id", prev.ConfirmToken); !errors.Is(err, mcp.ErrConsentRequired) {
		t.Fatalf("expired = %v", err)
	}

	prev, err = mcp.PreviewShare(ctx, db, "ada-id", clean.RecordID, "pool", "", []string{"productivity"}, []string{"dock"})
	if err != nil {
		t.Fatal(err)
	}
	shared, err := mcp.ShareNote(ctx, db, "ada-id", prev.ConfirmToken)
	if err != nil || shared.Status != "live" || shared.Alias != "ada" || shared.ID == "" {
		t.Fatalf("share = %+v err=%v", shared, err)
	}
	mine, err := mcp.ListMyShares(ctx, db, "ada-id")
	if err != nil || len(mine) != 1 || mine[0].ID != shared.ID {
		t.Fatalf("ada shares = %+v err=%v", mine, err)
	}
	beaShares, err := mcp.ListMyShares(ctx, db, "bea-id")
	if err != nil || len(beaShares) != 0 {
		t.Fatalf("bea shares = %+v", beaShares)
	}

	if _, err := mcp.UpdateNote(ctx, db, "ada-id", clean.RecordID, "", "private rewrite after share", nil, 0); err != nil {
		t.Fatal(err)
	}
	got, err := mcp.GetPoolShare(ctx, db, "bea-id", shared.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got.Content, "private rewrite") || strings.Contains(got.Content, "ada@example.com") || strings.Contains(got.Content, "ada-id") {
		t.Fatalf("get leaked: %+v", got)
	}
	if !strings.Contains(got.Content, `untrusted="true"`) || !strings.Contains(got.Attribution, "@ada") || !strings.Contains(got.Attribution, shared.ID) {
		t.Fatalf("get = %+v", got)
	}

	hits, err := mcp.SearchPool(ctx, db, "bea-id", "north gate", "", "", 10)
	if err != nil || len(hits) != 1 || hits[0].ID != shared.ID || hits[0].Alias != "ada" {
		t.Fatalf("search = %+v err=%v", hits, err)
	}
	blob := hits[0].Title + hits[0].Summary + hits[0].Alias
	if strings.Contains(blob, "ada@example.com") || strings.Contains(blob, "bea@example.com") || strings.Contains(blob, "ada-id") {
		t.Fatalf("search leaked email: %+v", hits)
	}

	copyRec, err := mcp.SavePoolCopy(ctx, db, "bea-id", shared.ID, mcp.KindNote)
	if err != nil || copyRec.Kind != mcp.KindNote {
		t.Fatalf("copy = %+v err=%v", copyRec, err)
	}
	if !strings.Contains(copyRec.Body, "@ada") || !strings.Contains(copyRec.Body, shared.ID) {
		t.Fatalf("copy missing attribution: %q", copyRec.Body)
	}
	if _, err := mcp.GetNote(ctx, db, "ada-id", copyRec.RecordID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("ada saw bea copy: %v", err)
	}

	if _, err := mcp.UnshareNote(ctx, db, "bea-id", shared.ID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("bea unshare = %v", err)
	}
	if _, err := mcp.GetPoolShare(ctx, db, "bea-id", shared.ID); err != nil {
		t.Fatalf("bea unshare hid ada: %v", err)
	}

	secret, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "keys", Body: "key sk-abcdefghijklmnopqrstuv"})
	if err != nil {
		t.Fatal(err)
	}
	secPrev, err := mcp.PreviewShare(ctx, db, "ada-id", secret.RecordID, "pool", "", nil, nil)
	if err != nil || len(secPrev.Findings) == 0 {
		t.Fatalf("secret preview = %+v err=%v", secPrev, err)
	}
	if _, err := mcp.ShareNote(ctx, db, "ada-id", secPrev.ConfirmToken); !errors.Is(err, mcp.ErrShareBlocked) {
		t.Fatalf("secret share = %v", err)
	}

	mail, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "mail", Body: "write other@example.com later"})
	if err != nil {
		t.Fatal(err)
	}
	mailPrev, err := mcp.PreviewShare(ctx, db, "ada-id", mail.RecordID, "pool", "", nil, nil)
	if err != nil || len(mailPrev.Findings) == 0 {
		t.Fatalf("mail preview = %+v err=%v", mailPrev, err)
	}
	if _, err := mcp.ShareNote(ctx, db, "ada-id", mailPrev.ConfirmToken); !errors.Is(err, mcp.ErrShareBlocked) {
		t.Fatalf("mail share = %v", err)
	}
	if n := shareCount(t, db); n != 1 {
		t.Fatalf("blocked shares stored %d", n)
	}

	stick, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Kind: mcp.KindStick, Title: "west watch", Body: "stand at the west gate"})
	if err != nil {
		t.Fatal(err)
	}
	stPrev, err := mcp.PreviewShare(ctx, db, "ada-id", stick.RecordID, "pool", "", []string{"tools-and-workflows"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	stShare, err := mcp.ShareNote(ctx, db, "ada-id", stPrev.ConfirmToken)
	if err != nil || stShare.Kind != mcp.KindStick {
		t.Fatalf("stick share = %+v err=%v", stShare, err)
	}
	stGot, err := mcp.GetPoolShare(ctx, db, "bea-id", stShare.ID)
	if err != nil || stGot.Kind != mcp.KindStick || !strings.Contains(stGot.Content, "west gate") {
		t.Fatalf("stick get = %+v err=%v", stGot, err)
	}

	linkNote, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "quiet link", Body: "only get by id sees this snapshot"})
	if err != nil {
		t.Fatal(err)
	}
	linkPrev, err := mcp.PreviewShare(ctx, db, "ada-id", linkNote.RecordID, "link", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	linkShare, err := mcp.ShareNote(ctx, db, "ada-id", linkPrev.ConfirmToken)
	if err != nil || linkShare.Visibility != "link" {
		t.Fatalf("link share = %+v err=%v", linkShare, err)
	}
	linkHits, err := mcp.SearchPool(ctx, db, "bea-id", "quiet", "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range linkHits {
		if h.ID == linkShare.ID {
			t.Fatalf("link share in search: %+v", h)
		}
	}
	if _, err := mcp.GetPoolShare(ctx, db, "bea-id", linkShare.ID); err != nil {
		t.Fatalf("link get = %v", err)
	}

	if _, err := mcp.UnshareNote(ctx, db, "ada-id", shared.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := mcp.GetPoolShare(ctx, db, "bea-id", shared.ID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("revoked get = %v", err)
	}
	after, err := mcp.SearchPool(ctx, db, "bea-id", "north gate", "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range after {
		if h.ID == shared.ID {
			t.Fatalf("revoked still indexed: %+v", h)
		}
	}

	hidePrev, err := mcp.PreviewShare(ctx, db, "ada-id", clean.RecordID, "pool", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	hideShare, err := mcp.ShareNote(ctx, db, "ada-id", hidePrev.ConfirmToken)
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"bea-id", "cara-id", "dan-id"} {
		if _, err := mcp.ReportPoolShare(ctx, db, user, hideShare.ID, "spam"); err != nil {
			t.Fatalf("report %s: %v", user, err)
		}
	}
	if _, err := mcp.GetPoolShare(ctx, db, "bea-id", hideShare.ID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("hidden get = %v", err)
	}

	topics, err := mcp.ListPoolTopics(ctx, db)
	if err != nil || len(topics) < 8 {
		t.Fatalf("topics = %+v err=%v", topics, err)
	}

	id := mcp.Identity{UserID: "ada-id"}
	if err := mcp.RequireScope(id, auth.ScopePoolPublish); !errors.Is(err, mcp.ErrForbiddenScope) {
		t.Fatalf("empty scopes granted publish: %v", err)
	}

	mcp.ResetPoolForTest()
	ratePrev, err := mcp.PreviewShare(ctx, db, "ada-id", clean.RecordID, "pool", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if !mcp.AllowRate("ada-id", "share_note") {
			t.Fatalf("prefill %d denied", i)
		}
	}
	if _, err := mcp.ShareNote(ctx, db, "ada-id", ratePrev.ConfirmToken); !errors.Is(err, mcp.ErrRateLimited) {
		t.Fatalf("rate = %v", err)
	}
}

func shareCount(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pool_share`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
