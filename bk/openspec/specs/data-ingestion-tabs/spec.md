# Data Ingestion Tabs

## Purpose

Define the unified data ingestion experience within the RAG Manager page, where crawl, API request, and AI gathering become alternative ingestion methods alongside manual entry and file upload.

## Requirements

### Requirement: 
The RAG Manager 'Add Data' dialog MUST include tabs for Manual Entry, File Upload, Web Crawl, API Request, and Research Gather — making all five ingestion paths available in one place.

#### Scenario: Dialog has five tabs

- **THEN** a user opens the Add Data dialog in RAG Manager, they see tabs: Manual Entry, File Upload, Web Crawl, API Request, Research Gather, the user can switch between them freely, and content specific to each mode is rendered within the dialog.

#### Scenario: Web Crawl tab saves to RAG

- **THEN** a user provides a URL, clicks Crawl & Save, the crawl runs, and the extracted content is saved to a RAG collection, and a success/error message is shown.

#### Scenario: API Request tab saves to RAG

- **THEN** a user configures a request (URL, method, headers, body), clicks Send & Save, the request executes, and the response is saved to a RAG collection.

#### Scenario: Research Gather tab saves to RAG

- **THEN** a user enters a research topic, clicks Gather & Save, the AI gathers data from Wikipedia/Reddit/web, and the synthesized report is saved to a RAG collection.

#### Scenario: Existing modes unchanged

- **THEN** a user selects Manual Entry tab, they see the same form as before (collection name, format, content textarea), submitting adds data directly. When they select File Upload tab, the same file-select UI appears.

### Requirement: 
After the ingestion tabs are in place, the SourcesHub page, Crawler.js, and Gathering.js MUST be deleted. The /sources route MUST redirect to /rag.

#### Scenario: Sources module removed

- **THEN** the refactor is complete, navigating to /sources redirects to /rag, the Header nav has no Sources entry, and SourcesHub.js, Crawler.js, and Gathering.js no longer exist.

### Requirement: 
Backend endpoints for crawl, gathering, and request execution MUST remain unchanged; only the frontend UI moves.

#### Scenario: Backend unchanged

- **THEN** API calls to /crawler/, /crawler/profiles, /gathering/, and /request-tools/ continue to work, the backend has no changes in this change.
