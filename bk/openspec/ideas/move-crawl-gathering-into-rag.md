# Move crawl and gathering into RAG Manager, remove Sources module

## Source
- Origin: human
- Created: 2026-08-06T14:36:15Z
- Tags: refactor, modules, sources, rag

## Prompt
The Sources module (SourcesHub, Crawler.js, Gathering.js) is being slimmed down. The crawl and request functionalities should become data ingestion tabs within the RAG Manager (RAGManager.js), alongside the existing manual entry and file upload modes. Gathering (AI-powered research) should also become a way to ingest data into RAG. After the move, delete the SourcesHub page and its sub-pages (Crawler.js, Gathering.js), update routes in App.js and Header.js.

<!-- OPENSPEC_IDEA_ENRICHMENT_START -->
## Enrichment Report

Generated: 2026-08-06T14:36:27Z

### Problem
TBD

### Proposed Direction
TBD

### Key Questions
- None

### Feasibility
Feasibility: High

### T-Shirt Size
T-Shirt Size: M

### Size Justification
RAGManager.js already has a dialog for Add Data with manual and file modes; crawler already saves to RAG collections (collection_name field in profiles), gathering already has 'Add to RAG' modal. Wiring is largely done — this is about moving UI tabs into the RAG page and deleting the old pages.

### Risks
- RAGManager.js is 720 lines already; adding three more ingestion modes makes it large — consider sub-components or collapsible sections
- API endpoints (/crawler/, /gathering/, /request-tools/) stay in backend; only frontend moves

### Suggested Next Step
Create OpenSpec change with spec defining new RAG Manager tabs: Manual, File Upload, Web Crawl, API Request, Research Gather
<!-- OPENSPEC_IDEA_ENRICHMENT_END -->
