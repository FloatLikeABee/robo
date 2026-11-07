# Spec Delta

## Purpose

Makes twilight cards, chat bubbles, and card titles read as separate objects from the page behind them.

## ADDED Requirements

### Requirement: Cards are lighter than the page
A card, chat bubble, or titled panel on a twilight page SHALL use a fill lighter than the page floor `#2c2640` and a visible edge. The page floor SHALL stay `#2c2640`.

#### Scenario: Assistant chat bubble
- **WHEN** an operator views an assistant message on Dorph AI
- **THEN** the bubble fill is lighter than the page behind it
- **AND** the bubble has a visible border

#### Scenario: App card
- **WHEN** an operator views a card in DorphNotes, Content, Event Logs, Skills, or DorphTools
- **THEN** that card fill is lighter than the page behind it

### Requirement: Card titles are brighter than card body text
The title of a card SHALL use lighter or accent ink than the secondary text on that same card.

#### Scenario: Titled card
- **WHEN** a card shows a title and supporting text
- **THEN** the title is brighter than the supporting text

### Requirement: User bubbles stay distinct
A user chat bubble SHALL stay the stronger purple fill so it does not match the assistant card.

#### Scenario: Two bubbles
- **WHEN** a user message and an assistant message are both on screen
- **THEN** the user bubble is a deeper purple than the assistant card
