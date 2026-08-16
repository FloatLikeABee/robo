# Move crawl + gathering into RAG Manager, remove Sources module

## Motivation

Slimming down modules. Crawl, API requests, and AI research gathering are all ways to get data into RAG collections — not standalone modules. Moving them into the RAG Manager's "Add Data" dialog as alternative ingestion tabs creates a single entry point for all data ingestion.

## Approach

1. Refactor `RAGManager.js`:
   - Convert the existing "Add Data" dialog from a flat form into a **tabbed dialog** with 5 tabs:
     - **Manual Entry** (existing manual mode — no change)
     - **File Upload** (existing file upload — no change)
     - **Web Crawl** (lifted from Crawler.js's WebCrawlersPanel)
     - **API Request** (lifted from Crawler.js's RequestToolsPanel)
     - **Research Gather** (lifted from Gathering.js)
   - Each tab embeds the relevant form/UI inline.
   - Web Crawl and API Request tabs reuse existing backend endpoints (`/crawler/`, `/request-tools/`).
   - Research Gather tab calls `/gathering/` then offers "Save to RAG" within the same flow.

2. Delete:
   - `frontend/src/pages/SourcesHub.js`
   - `frontend/src/pages/Crawler.js`
   - `frontend/src/pages/Gathering.js`

3. Update routes:
   - `App.js`: `/sources` → redirect to `/rag-tools`; remove SourcesHub import
   - `Header.js`: remove "Sources" nav item

## Backend

No backend changes. All existing endpoints remain untouched.