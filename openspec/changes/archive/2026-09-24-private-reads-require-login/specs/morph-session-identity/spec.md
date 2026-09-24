## ADDED Requirements

### Requirement: Production tokens cannot be issued in the future
When Morph is in production mode, decoding a JWT MUST reject a token whose `iat` is later than now plus a small clock-skew leeway. A token whose `iat` is within that leeway MUST still be accepted when its other production lifetime checks pass. Local mode MUST keep accepting a token that omits `iat`.

#### Scenario: Future iat is rejected in production
- **WHEN** production mode decodes a token whose `iat` is more than the clock-skew leeway ahead of now
- **THEN** decoding fails

#### Scenario: iat inside the leeway is accepted in production
- **WHEN** production mode decodes a token whose `iat` is within the clock-skew leeway of now and whose lifetime is within the production maximum
- **THEN** decoding succeeds

#### Scenario: Local mode still accepts a token without iat
- **WHEN** local mode decodes a token that has `exp` and no `iat`
- **THEN** decoding succeeds

### Requirement: Chat does not assume admin when the user id is missing
A chat request whose resolved user id is empty MUST fail closed with HTTP 401. It MUST NOT proceed as user id `admin`.

#### Scenario: Empty user id is not admin
- **WHEN** the chat handler runs with no user id
- **THEN** the response status is 401
- **AND** the handler does not record the exchange as user `admin`
