## Purpose

Notes & TODOs AI assist drafts and polishes personal notes and todos from the user’s title and body, without leftover school-transportation domain.

## ADDED Requirements

### Requirement: Generate from the user’s title

When the user runs AI assist on a Notes & TODOs item with a title and little or no body, the system MUST send that title to the model as the topic. The returned body MUST be about that title. The prompt MUST NOT cast the user as transportation or school-operations staff, and MUST NOT use school-bus, student, parent-portal, semester-schedule, or similar leftover examples unless those words appear in the user’s title (or body, when improving).

#### Scenario: Todo titled make money on AI stock

- **WHEN** the user sets a todo title to “make money on AI stock” (or equivalent) and runs AI assist with an empty or short body
- **THEN** the filled body is about that title (AI / stocks / making money)
- **AND** the body MUST NOT be school-bus routing, student addresses, driver assignments, parent pickup times, or other leftover Skool operations copy

#### Scenario: Note generate uses the title

- **WHEN** the user sets a note title and runs AI assist with an empty or short body
- **THEN** the filled body is about that title
- **AND** the prompt MUST NOT instruct a transportation or school-operations persona

### Requirement: Empty title does not invent school operations

When AI assist generate runs with no title/seed, the system MUST NOT pick a school-transportation example topic (route change, vehicle check, parent communication, students, schedules). It MAY write a generic personal note or todo, or require a title in the UI (title-first, as today).

#### Scenario: No leftover default topic

- **WHEN** generate runs with an empty seed
- **THEN** the prompt MUST NOT include leftover school-operations example topics

### Requirement: Improve keeps the user’s topic

When AI assist improves existing body text, the system MUST keep the same meaning. If a title is present, it MUST be passed as additional context so polish cannot drift into leftover school-ops copy.

#### Scenario: Improve with a title

- **WHEN** the user has a title and a body of 20 or more characters and runs AI assist
- **THEN** the request includes that title as context
- **AND** the improved body stays on the same topic as the original body and title
