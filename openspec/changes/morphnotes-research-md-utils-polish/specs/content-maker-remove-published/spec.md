## Purpose

Content Maker Published contents lets operators remove a published page from the list so it is no longer public, not only delete unpublished Saved drafts.

## ADDED Requirements

### Requirement: Published rows can be removed
Each Published contents row MUST offer a remove action, including rows whose status is Published. Removing a published page MUST delete that published record so it no longer appears in the list.

#### Scenario: Remove a published-only page
- **WHEN** the operator removes a Published contents row that is Published and has no Saved draft
- **THEN** the row disappears from the list
- **AND** the page is no longer listed as published

#### Scenario: Saved drafts still removable
- **WHEN** the operator removes a Saved-only row
- **THEN** that draft is deleted
- **AND** the row disappears from the list

### Requirement: Public URL stops serving after remove
After a published page is removed, its previous public path MUST NOT serve that HTML to anonymous visitors (not found or gone). Republishing the same name MAY mint or reuse a slug per existing publish-once rules.

#### Scenario: Open after remove
- **WHEN** the operator has removed a published page whose public path was `/public/p/{slug}`
- **THEN** requesting that public path does not return the old HTML document

### Requirement: Confirm before remove
Remove MUST ask for confirmation. Cancel on the confirmation MUST leave the published page and list unchanged.

#### Scenario: Confirm removal
- **WHEN** the operator chooses Remove and confirms
- **THEN** the published or saved item is deleted as specified above

#### Scenario: Cancel confirmation
- **WHEN** the operator chooses Remove then cancels the confirmation
- **THEN** the list still contains that row
- **AND** a published URL still serves if it did before

### Requirement: Merged published-plus-draft row can leave the list
If a row represents both a Saved draft and a Published page, Remove MUST take the row off the Published contents list (delete the published page and the linked draft for that name).

#### Scenario: Remove merged row
- **WHEN** the operator removes a row that has both a Saved draft and a Published page for the same name
- **THEN** that name no longer appears in Published contents
- **AND** the public path for that page no longer serves the old HTML
