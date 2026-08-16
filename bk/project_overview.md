# Project Overview: Ground Control with MCP Support

## Project Name
Ground Control with MCP Support

## Description
A comprehensive Retrieval-Augmented Generation (RAG) system with Model Context Protocol (MCP) support for Ollama models. This system provides a complete solution for building, managing, and deploying AI agents with RAG capabilities.

## Key Features

### 🧠 Ground Control
- **Vector Database**: ChromaDB integration for efficient document storage and retrieval
- **Multiple Formats**: Support for JSON, CSV, and text data formats
- **Data Validation**: Built-in validation for all data inputs
- **Collection Management**: Create, query, and manage RAG collections
- **Semantic Search**: Advanced embedding-based search capabilities

### 🤖 Agent Management
- **Configurable Agents**: Create agents with different types (RAG, Tool, Hybrid)
- **Model Support**: Integration with any Ollama model
- **Tool Integration**: Connect agents with various tools
- **Real-time Execution**: Run agents and get responses through the API

### 🛠️ Tool System
- **Email Tools**: Send emails with SMTP configuration
- **Web Search**: DuckDuckGo integration for web searches
- **Calculator**: Mathematical computation capabilities
- **Financial Data**: Real-time stock prices and financial data (yfinance/Yahoo Finance)
- **Wikipedia**: Wikipedia search integration
- **Browser Automation**: AI-powered browser control using LangChain + Playwright
- **Custom Tools**: Extensible tool system for custom functionality

### 🔌 MCP (Model Context Protocol)
- **Stable Communication**: Reliable protocol for AI model communication
- **Real-time Updates**: Live status monitoring and updates
- **Client Support**: Multiple client connections
- **Message Broadcasting**: Broadcast messages to all connected clients

### 🎨 Modern UI
- **React + Material-UI**: Beautiful, responsive web interface
- **Real-time Updates**: Live system status and monitoring
- **Intuitive Design**: Easy-to-use interface for all operations
- **Mobile Responsive**: Works on all device sizes

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Frontend│    │   FastAPI Backend│    │   Ollama Models │
│                 │◄──►│                 │◄──►│                 │
│  Material-UI    │    │  Ground Control │    │  Local/Remote   │
│  Components     │    │  Agent Manager  │    │  AI Models      │
└─────────────────┘    │  Tool Manager   │    └─────────────────┘
                       │  MCP Service    │
                       └─────────────────┘
                                │
                       ┌─────────────────┐
                       │   ChromaDB      │
                       │   Vector Store  │
                       └─────────────────┘
```

## Technology Stack

### Backend
- **Framework**: FastAPI
- **Database**: ChromaDB (vector database)
- **Embeddings**: HuggingFace sentence transformers
- **LLM Integration**: LangChain with support for multiple providers (Gemini, Qwen, Mistral, Groq)
- **Data Processing**: Pandas, JSON, CSV parsing
- **Tools**: Playwright for browser automation

### Frontend
- **Framework**: React with Material-UI
- **Responsive Design**: Mobile-friendly interface

### Key Libraries/Tools
- LangChain ecosystem for LLM integration
- ChromaDB for vector storage
- HuggingFace Hub for embeddings and models
- Playwright for browser automation
- Pydantic for data validation and API models
- Uvicorn for ASGI server

## Project Structure

```
.
├── README.md
├── requirements.txt
├── main.py                 # Main entry point
├── src/                    # Source code
│   ├── api.py              # FastAPI routes
│   ├── config.py           # Configuration loading
│   ├── models.py           # Pydantic models
│   ├── rag_system.py       # RAG core functionality
│   ├── llm_langchain_wrapper.py  # LLM wrapper
│   ├── llm_factory.py      # LLM provider factory
│   ├── tools.py            # Tool system
│   └── ...
├── env.example             # Environment configuration template
├── chroma_db/              # ChromaDB persistent storage
├── data/                   # Data storage directory
└── frontend/              # React frontend
```

## Core Components

### RAG System (src/rag_system.py)
- Vector database integration with ChromaDB
- Data validation for multiple formats (JSON, CSV, TXT)
- Text splitting and embedding using sentence transformers
- Collection management (create, query, delete)
- Smart Import functionality that processes data with AI

### LLM Integration (src/llm_factory.py, src/llm_langchain_wrapper.py)
- Support for multiple LLM providers (Gemini, Qwen, Mistral, Groq)
- Integration with LangChain ecosystem
- Configurable model selection and parameters

### Tools System (src/tools.py)
- Extensible tool framework with built-in tools for email, web search, calculator, financial data, etc.
- Browser automation tool leveraging Playwright
- Custom tool support

## API Endpoints

### System
- `GET /status` - Get system status
- `GET /models` - List available models

### RAG
- `GET /rag/collections` - List collections
- `POST /rag/collections/{name}/data` - Add data
- `POST /rag/collections/{name}/query` - Query collection
- `DELETE /rag/collections/{name}` - Delete collection
- `POST /rag/validate` - Validate data

### Agents
- `GET /agents` - List agents
- `POST /agents` - Create agent
- `GET /agents/{id}` - Get agent details
- `PUT /agents/{id}` - Update agent
- `DELETE /agents/{id}` - Delete agent
- `POST /agents/{id}/run` - Run agent

### Tools
- `GET /tools` - List tools
- `PUT /tools/{id}` - Update tool config

### MCP
- `POST /mcp/start` - Start MCP server

## Configuration

The system uses an environment-based configuration system:
- Config file: `.env` or `env` at project root
- Settings loaded via Pydantic Settings
- Supports different LLM providers (Gemini, Qwen, Mistral, Groq)
- Database and embedding configurations
- CORS settings
- Email and financial API configurations

## Installation Requirements

### Prerequisites
1. Python 3.13+
2. Node.js 16+
3. Ollama (for local AI models)
4. Git

### Python Dependencies (requirements.txt)
- FastAPI, uvicorn, pydantic
- ChromaDB, sentence-transformers
- LangChain and related libraries
- Playwright for browser automation
- Various utility packages for data processing

## Quick Start Steps

1. Clone repository
2. Install Python dependencies: `pip install -r requirements.txt`
3. Install Playwright browsers: `playwright install`
4. Install frontend dependencies: `cd frontend && npm install`
5. Set up environment variables: `cp .env.example .env`
6. Start Ollama server
7. Start backend: `python main.py`
8. Start frontend: `cd frontend && npm start`
9. Access application at http://localhost:3000

## Usage Patterns

### RAG Data Management
1. Navigate to RAG Manager
2. Click "Add Data"
3. Choose data format (JSON, CSV, or Text)
4. Enter data content
5. Add tags and metadata
6. Submit to create a collection
7. Query collections with semantic search

### Agent Creation
1. Go to Agent Manager
2. Click "Create Agent"
3. Configure settings (name, description, type, model, temperature, etc.)
4. Select RAG collections and tools to use
5. Save the agent
6. Run agent with natural language queries

### Tool Configuration
1. Navigate to Tool Manager
2. View all available tools
3. Configure tools as needed
4. Tools can be enabled/disabled for specific agents

### Browser Automation
The system includes a Browser Automation tool that uses AI (LangChain) and Playwright to control a web browser and follow natural language instructions.