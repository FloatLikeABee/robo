---
slug: event-rsvp
title: Event RSVP survey
tags: [event, rsvp]
---

# Instructions
Collect RSVP details for an upcoming event.

## Q1 — Attendee name
- field: attendee_name
- collect: text
- required: true
- prompt: What is the attendee's full name?

## Q2 — Attendance
- field: attendance
- collect: mcp_html
- widget: select
- options: [Yes, No, Maybe]
- required: true
- prompt: Will they attend?

## Q3 — Guest count
- field: guest_count
- collect: text
- required: false
- prompt: How many guests (including themselves)?

## Q4 — Dietary notes
- field: dietary
- collect: text
- required: false
- prompt: Any dietary restrictions or notes?
