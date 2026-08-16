## Purpose

Defines Morph login persistence so authenticated users stay signed in until they explicitly sign out, without forced session timeout.

## ADDED Requirements

### Requirement: No session timeout after login
The system SHALL keep a successful Morph login valid until the user explicitly signs out. The system MUST NOT expire the session solely due to elapsed time, idle time, or a fixed cookie Max-Age timeout used as a logout policy.

#### Scenario: User remains signed in after long idle
- **WHEN** a user has successfully signed in and later returns after a long idle period without signing out
- **THEN** the user remains authenticated and can use Morph without being forced to log in again solely because of timeout

#### Scenario: Explicit sign-out still clears session
- **WHEN** a signed-in user chooses Sign out
- **THEN** the session is cleared and protected Morph surfaces require login again
