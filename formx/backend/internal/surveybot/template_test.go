package surveybot

import "testing"

func TestParseMarkdownOnboarding(t *testing.T) {
	md := `---
slug: staff-onboarding
title: Staff onboarding survey
tags: [onboarding, hr]
---

# Instructions
Ask one at a time.

## Q1 — Full name
- field: name
- collect: text
- required: true
- prompt: What is your full name?

## Q2 — Gender
- field: gender
- collect: mcp_html
- widget: select
- options: [Female, Male, Non-binary]
- required: true
- prompt: Please select your gender.
`
	p, err := ParseMarkdown(md)
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Staff onboarding survey" {
		t.Fatalf("title=%q", p.Title)
	}
	if len(p.Steps) != 2 {
		t.Fatalf("steps=%d", len(p.Steps))
	}
	if p.Steps[1].Collect != "mcp_html" || len(p.Steps[1].Options) != 3 {
		t.Fatalf("step2=%+v", p.Steps[1])
	}
	block := BlockForStep(p.Steps[1])
	if block == nil || block.Widget != "select" {
		t.Fatalf("block=%+v", block)
	}
}

func TestParseSurveyAnswerMessage(t *testing.T) {
	f, v, ok := ParseSurveyAnswerMessage("survey_bot_answer:gender=female")
	if !ok || f != "gender" || v != "female" {
		t.Fatalf("%v %q %q", ok, f, v)
	}
}

func TestParseDescriptionOnly(t *testing.T) {
	md := `---
slug: weekend-parties
title: Weekend party preference survey
---

This is a preference survey for weekend parties. We want to know what staff wants to do every weekend.
`
	p, err := ParseMarkdown(md)
	if err != nil {
		t.Fatal(err)
	}
	if !p.NeedsCompile {
		t.Fatal("expected NeedsCompile")
	}
	if len(p.Steps) != 0 {
		t.Fatalf("steps=%d", len(p.Steps))
	}
	if p.Title != "Weekend party preference survey" {
		t.Fatalf("title=%q", p.Title)
	}
}
