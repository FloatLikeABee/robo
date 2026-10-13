package grill

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const exampleTitle = "make this morphAI a platform for agents, actually agent notebook"

const chattyReply = "Great! I'm Morph AI, designed to help developers create and manage AI agents. If you have any requests or need assistance with setting up an agent notebook, feel free to let me know."

func TestQuestionCapFollowsTheTitle(t *testing.T) {
	if got := QuestionCap("buy milk"); got != 1 {
		t.Fatalf("short title cap = %d", got)
	}
	if got := QuestionCap(exampleTitle); got < 2 || got > MaxQuestions {
		t.Fatalf("example cap = %d", got)
	}
	long := strings.Repeat("roadmap ", 30) + ", and sharing, and migration"
	if got := QuestionCap(long); got != MaxQuestions {
		t.Fatalf("long title cap = %d", got)
	}
	if got := QuestionCap("buy milk"); got != QuestionCap("buy milk") {
		t.Fatal("cap is not stable")
	}
}

func TestQuestionCapIgnoresBodyLength(t *testing.T) {
	store := NewStore()
	body := strings.Repeat("existing paragraph. ", 80)
	view, err := store.Act(context.Background(), "ada", Input{
		Action: "start", Kind: KindNote, Title: "buy milk", Description: body,
	}, func(context.Context, string) (string, error) {
		return "What kind of milk?\nWho is it for?\nWhen do you need it?\n", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if view.Planned != 1 || !strings.Contains(strings.ToLower(view.Question), "milk") {
		t.Fatalf("view = %+v", view)
	}
}

func TestParseQuestionsDropsChatAndStaysOnTitle(t *testing.T) {
	raw := "Great! I'm Morph AI.\nWhat is your favorite color?\nWhat should the agent notebook include?"
	got := ParseQuestions(raw, KindNote, exampleTitle, 3)
	if len(got) != 3 {
		t.Fatalf("questions = %#v", got)
	}
	if !strings.Contains(strings.ToLower(got[0]), "notebook") && !strings.Contains(strings.ToLower(got[0]), "agent") {
		t.Fatalf("first = %q", got[0])
	}
	for _, q := range got {
		if strings.Contains(q, "Great") || strings.Contains(strings.ToLower(q), "favorite color") {
			t.Fatalf("off topic %q in %#v", q, got)
		}
	}
}

func TestPromptsDoNotAskTheModelToAnswer(t *testing.T) {
	plan := PlanPrompt(KindNote, exampleTitle, "", 3)
	write := WritePrompt(Session{Kind: KindNote, Title: exampleTitle, Turns: []Turn{{Question: "Who is it for?", Answer: "Agents"}}})
	for _, p := range []string{plan, write} {
		if strings.Contains(p, "Answer the question") || strings.Contains(p, "I'm Morph") {
			t.Fatalf("prompt invites a chat reply:\n%s", p)
		}
	}
	if !strings.Contains(write, "Agents") || !strings.Contains(write, exampleTitle) {
		t.Fatal("write prompt dropped the interview")
	}
	if !strings.Contains(plan, "exactly 3") {
		t.Fatal(plan)
	}
}

func TestWritePromptLoadsHTMLSkillAndStaysBodyOnly(t *testing.T) {
	write := WritePrompt(Session{Kind: KindNote, Title: exampleTitle, Turns: []Turn{{Question: "Who is it for?", Answer: "Agents"}}})
	for _, needle := range []string{"html-anything", "Arial", "#38bdf8", "Output only the body."} {
		if !strings.Contains(write, needle) {
			t.Fatalf("missing %q in:\n%s", needle, write)
		}
	}
	if strings.Contains(strings.ToLower(write), "cyan #38bdf8") {
		t.Fatal("write prompt still names cyan before the hex")
	}
	if !strings.Contains(strings.ToLower(write), "not a letter") && !strings.Contains(strings.ToLower(write), "do not write a letter") {
		t.Fatal("write prompt must forbid a letter")
	}
}

func TestQuestionCapCoversTitleParts(t *testing.T) {
	got := QuestionCap("Team offsite plan for 12 people in Hangzhou")
	if got < 3 || got > MaxQuestions {
		t.Fatalf("3-part title cap = %d", got)
	}
	if got := QuestionCap("buy milk"); got != 1 {
		t.Fatalf("short title cap = %d", got)
	}
}

const letterHTML = `<!DOCTYPE html>
<html>
<head>
<style>
h1, h2, h3 { color: cyan #38bdf8; }
</style>
</head>
<body>
<h1>Team Offsite Plan</h1>
<p>Dear Team,</p>
<p>We've selected West Lake.</p>
<p>Best regards,</p>
<p>[Your Name]</p>
</body>
</html>`

func TestSanitizeDropsLetterHTML(t *testing.T) {
	sess := Session{
		Kind:  KindNote,
		Title: "Team offsite plan for 12 people in Hangzhou",
		Turns: []Turn{{Question: "Where should it be?", Answer: "West Lake area"}},
	}
	got := SanitizeDraft(sess, letterHTML)
	low := strings.ToLower(got)
	if strings.Contains(low, "dear team") || strings.Contains(got, "[Your Name]") || strings.Contains(low, "<!doctype html") {
		t.Fatalf("letter leaked: %q", got)
	}
	if !strings.Contains(got, "West Lake area") {
		t.Fatalf("draft dropped the answer: %q", got)
	}
}

func TestStopPhraseIsTheWholeAnswer(t *testing.T) {
	if !IsStopPhrase("enough, write it now") || !IsStopPhrase("  DONE ") {
		t.Fatal("expected stop phrases")
	}
	if IsStopPhrase("enough detail about the API, please write the risks") {
		t.Fatal("a real answer was treated as a stop phrase")
	}
}

func TestSanitizeReplacesTheChattyScreenshot(t *testing.T) {
	sess := Session{
		Kind:  KindNote,
		Title: exampleTitle,
		Turns: []Turn{{Question: "Who is it for?", Answer: "People building agents"}},
	}
	got := SanitizeDraft(sess, chattyReply)
	if strings.Contains(got, "I'm Morph") || strings.Contains(strings.ToLower(got), "feel free") {
		t.Fatalf("draft = %q", got)
	}
	if !strings.Contains(got, "People building agents") || !strings.Contains(got, exampleTitle) {
		t.Fatalf("draft = %q", got)
	}
}

func TestSanitizeKeepsARealNote(t *testing.T) {
	sess := Session{Kind: KindNote, Title: "buy milk"}
	note := "Whole milk, two bottles.\n\nLeave it in the fridge."
	if got := SanitizeDraft(sess, "```\n"+note+"\n```"); got != note {
		t.Fatalf("got %q", got)
	}
}

func TestInterviewSkipFinishAndOwner(t *testing.T) {
	store := NewStore()
	var writes int
	complete := func(_ context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "Write the body") {
			writes++
			return chattyReply, nil
		}
		return "What should the agent notebook include?\nWho is the first user?\nWhat is out of scope?\n", nil
	}
	started, err := store.Act(context.Background(), "ada", Input{Action: "start", Kind: KindNote, Title: exampleTitle}, complete)
	if err != nil {
		t.Fatal(err)
	}
	if started.Phase != PhaseAsking || started.Planned < 2 || started.Question == "" {
		t.Fatalf("start = %+v", started)
	}
	if _, err := store.Act(context.Background(), "bea", Input{Action: "answer", ID: started.ID, Answer: "stolen"}, complete); !errors.Is(err, ErrNotFound) {
		t.Fatalf("bea err = %v", err)
	}
	skipped, err := store.Act(context.Background(), "ada", Input{Action: "skip", ID: started.ID}, complete)
	if err != nil || skipped.Draft != "" || writes != 0 {
		t.Fatalf("skip = %+v writes=%d err=%v", skipped, writes, err)
	}
	if _, err := store.Act(context.Background(), "ada", Input{Action: "answer", ID: started.ID, Answer: "People building agents"}, complete); err != nil {
		t.Fatal(err)
	}
	finished, err := store.Act(context.Background(), "ada", Input{Action: "answer", ID: started.ID, Answer: "enough, write it now"}, complete)
	if err != nil {
		t.Fatal(err)
	}
	if finished.Phase != PhaseReview || strings.Contains(finished.Draft, "I'm Morph") {
		t.Fatalf("finish = %+v", finished)
	}
	if !strings.Contains(finished.Draft, "People building agents") {
		t.Fatalf("draft dropped the answer: %q", finished.Draft)
	}
	if strings.Contains(finished.Draft, "enough, write it now") {
		t.Fatal("stop phrase was stored as a fact")
	}
	if _, err := store.Act(context.Background(), "bea", Input{Action: "finish", ID: started.ID}, complete); !errors.Is(err, ErrNotFound) {
		t.Fatalf("bea finish = %v", err)
	}
}

func TestEmptyAnswerDoesNotAdvance(t *testing.T) {
	store := NewStore()
	started, err := store.Act(context.Background(), "ada", Input{Action: "start", Kind: KindTodo, Title: "buy milk"}, func(context.Context, string) (string, error) {
		return "What kind of milk?", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Act(context.Background(), "ada", Input{Action: "answer", ID: started.ID, Answer: "  "}, nil)
	if !errors.Is(err, ErrAnswerRequired) {
		t.Fatalf("err = %v", err)
	}
	again, err := store.Act(context.Background(), "ada", Input{Action: "skip", ID: started.ID}, func(context.Context, string) (string, error) {
		return "- Buy milk", nil
	})
	if err != nil || again.Phase != PhaseReview || !strings.Contains(again.Draft, "milk") {
		t.Fatalf("after skip = %+v err=%v", again, err)
	}
}

func TestWriteFailureKeepsAnswers(t *testing.T) {
	store := NewStore()
	fail := errors.New("model down")
	calls := 0
	complete := func(_ context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "Write the body") {
			calls++
			if calls == 1 {
				return "", fail
			}
			return "A notebook for agents.", nil
		}
		return "What should the notebook include?", nil
	}
	started, err := store.Act(context.Background(), "ada", Input{Action: "start", Kind: KindNote, Title: "agent notebook"}, complete)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Act(context.Background(), "ada", Input{Action: "finish", ID: started.ID}, complete); !errors.Is(err, fail) {
		t.Fatalf("err = %v", err)
	}
	retried, err := store.Act(context.Background(), "ada", Input{Action: "finish", ID: started.ID}, complete)
	if err != nil || retried.Phase != PhaseReview || !strings.Contains(retried.Draft, "notebook") {
		t.Fatalf("retry = %+v err=%v", retried, err)
	}
}

func TestStartTimelineAndUnknownKind(t *testing.T) {
	store := NewStore()
	started, err := store.Act(context.Background(), "ada", Input{Action: "start", Kind: KindTimeline, Title: "shift log"}, func(context.Context, string) (string, error) {
		return "What happened on the shift log?", nil
	})
	if err != nil || started.Kind != KindTimeline || started.ID == "" {
		t.Fatalf("start = %+v err=%v", started, err)
	}
	_, err = store.Act(context.Background(), "ada", Input{Action: "start", Kind: "banana", Title: "nope"}, func(context.Context, string) (string, error) {
		t.Fatal("model called for unknown kind")
		return "", nil
	})
	if !errors.Is(err, ErrKind) {
		t.Fatalf("banana err = %v", err)
	}
}

func TestStartRequiresTitleAndDoesNotCallModel(t *testing.T) {
	called := false
	_, err := NewStore().Act(context.Background(), "ada", Input{Action: "start", Kind: KindNote, Title: "  "}, func(context.Context, string) (string, error) {
		called = true
		return "", nil
	})
	if !errors.Is(err, ErrTitleRequired) || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
}

func TestSessionCapDropsOldest(t *testing.T) {
	store := NewStore()
	var first string
	for i := 0; i < MaxPerOwner+1; i++ {
		view, err := store.Act(context.Background(), "ada", Input{Action: "start", Kind: KindNote, Title: "buy milk"}, func(context.Context, string) (string, error) {
			return "What kind of milk?", nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = view.ID
		}
	}
	_, err := store.Act(context.Background(), "ada", Input{Action: "skip", ID: first}, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("oldest err = %v", err)
	}
}
