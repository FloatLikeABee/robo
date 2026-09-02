## Purpose

Remove the now-unused Images and Readers modules from the AI Tools app to reduce surface area and maintenance, while preserving the core image-generation capability that other modules still depend on.

## ADDED Requirements

### Requirement: Images module is removed from AI Tools

The AI Tools app SHALL no longer expose the user-facing Images module. Its navigation entry, page/route, and image-management HTTP endpoints SHALL be removed.

#### Scenario: Images navigation is gone

- **WHEN** a user views the AI Tools navigation
- **THEN** there is no Images entry

#### Scenario: Images endpoints are gone

- **WHEN** a client requests a removed image-management endpoint
- **THEN** the request is not served by an Images route (it returns not found)

### Requirement: Core image generation is preserved for dependent modules

Removing the Images module SHALL NOT remove the underlying image-generation capability used by other modules. Graphic Documents and Video Stories SHALL continue to generate images as before.

#### Scenario: Graphic Documents still generate images

- **WHEN** a Graphic Document is generated
- **THEN** its embedded images are produced successfully

#### Scenario: Video Stories still generate images

- **WHEN** a Video Story scene image is generated
- **THEN** the image is produced successfully

### Requirement: Readers module is removed from AI Tools

The AI Tools app SHALL no longer expose the Readers module. The Image Reader and PDF Reader pages, the Readers route, its legacy redirect routes, and the reader HTTP endpoints SHALL be removed.

#### Scenario: Readers navigation is gone

- **WHEN** a user views the AI Tools navigation
- **THEN** there is no Readers entry

#### Scenario: Reader endpoints are gone

- **WHEN** a client requests a removed image-reader or pdf-reader endpoint
- **THEN** the request is not served by a Readers route (it returns not found)
