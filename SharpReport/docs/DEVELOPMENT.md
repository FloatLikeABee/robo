# DataPulse Development Guide

## Getting Started

### Prerequisites

- **Rust 1.84+**: [Install Rust](https://www.rust-lang.org/tools/install)
- **Node.js 22+**: [Install Node.js](https://nodejs.org/)
- **Java 17+**: Required for Metabase
- **Docker**: Optional, for containerized development

### Project Setup

```bash
# Clone the repository
git clone https://github.com/your-repo/datapulse.git
cd datapulse

# Set up environment
cp .env.example .env

# Install backend dependencies
cd backend
cargo build

# Install frontend dependencies
cd ../frontend
npm install
```

## Development Workflow

### Backend Development

```bash
# Start backend server with hot reload
cd backend
cargo run

# Run tests
cargo test

# Run linter
cargo clippy

# Format code
cargo fmt
```

### Frontend Development

```bash
# Start frontend dev server
cd frontend
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Full Stack Development

```bash
# In root directory
npm run dev
```

## Project Structure

```
datapulse/
├── backend/          # Rust backend
│   ├── src/          # Source code
│   ├── Cargo.toml    # Rust dependencies
│   └── config/       # Configuration files
├── frontend/         # Svelte frontend
│   ├── src/          # Source code
│   ├── package.json  # Node dependencies
│   └── public/       # Static assets
├── deploy/           # Deployment configs
├── docs/            # Documentation
├── design.md        # Architecture design
└── README.md         # Project overview
```

## Backend Architecture

### Key Components

- **Axum**: Web framework
- **SQLx**: Database ORM
- **Tokio**: Async runtime
- **Metabase Orchestrator**: Manages Metabase process
- **API Modules**: REST endpoints
- **Services**: Business logic

### Adding a New API Endpoint

1. Create a new file in `backend/src/api/`
2. Define your route handlers
3. Add the route to `main.rs`
4. Create service methods if needed

### Database Migrations

```bash
# Create a new migration
# Add SQL file to backend/migrations/

# Migrations run automatically on startup
```

## Frontend Architecture

### Key Components

- **Svelte 5**: Reactive UI framework
- **TailwindCSS**: Utility-first CSS
- **Lucide**: Icon library
- **ECharts**: Data visualization
- **Stores**: State management

### Adding a New Page

1. Create a new file in `frontend/src/routes/`
2. Define your Svelte component
3. Add to navigation if needed

### Adding a New Component

1. Create component in `frontend/src/lib/components/`
2. Use proper naming convention (PascalCase)
3. Add props and events as needed

## Common Tasks

### Authentication

```typescript
// Login
import { login } from '$lib/stores/auth.svelte';

const handleLogin = async (email, password) => {
    try {
        await login(email, password);
        // Redirect to protected page
    } catch (error) {
        // Handle error
    }
};
```

### API Calls

```typescript
// GET request
const response = await fetch('/api/v1/endpoint', {
    headers: {
        'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
    },
});

// POST request
const response = await fetch('/api/v1/endpoint', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
    },
    body: JSON.stringify(data),
});
```

### Theme Management

```typescript
import { theme, toggleTheme } from '$lib/stores/theme.svelte';

// Get current theme
const currentTheme = $theme;

// Toggle theme
<button on:click={toggleTheme}>Toggle Theme</button>
```

## Testing

### Backend Tests

```bash
# Run all tests
cargo test

# Run specific test
cargo test test_function_name

# Test with coverage
cargo tarpaulin
```

### Frontend Tests

```bash
# Run Svelte checks
npm run check

# Watch mode
npm run check:watch
```

## Debugging

### Backend Debugging

```bash
# Run with logging
RUST_LOG=debug cargo run

# Attach debugger (VS Code)
# Use Rust extension and launch configuration
```

### Frontend Debugging

```bash
# Run with debug mode
npm run dev

# Use browser dev tools
# Chrome/Firefox/Safari developer tools
```

## Metabase Integration

### Embedding Dashboards

```svelte
<script>
    import Embed from '$lib/components/Embed.svelte';
    
    let embedUrl = '/embed/dashboard/1';
</script>

<Embed {embedUrl} height="600px" />
```

### Metabase API

```rust
use crate::metabase::api_client::MetabaseApiClient;

let client = MetabaseApiClient::new("http://localhost:3001");
let dashboards = client.get_dashboards().await?;
```

## Deployment

### Local Development

```bash
# Start both frontend and backend
docker-compose -f deploy/docker-compose.dev.yml up
```

### Production Build

```bash
# Build frontend
cd frontend
npm run build

# Build backend
cd ../backend
cargo build --release

# Run with Docker
docker-compose -f deploy/docker-compose.prod.yml up -d
```

## Best Practices

### Code Style

- Follow Rust naming conventions
- Use Svelte best practices
- Keep components small and focused
- Use TypeScript for complex logic

### Security

- Never commit secrets
- Use environment variables
- Validate all inputs
- Use prepared statements
- Implement proper auth

### Performance

- Use lazy loading
- Cache API responses
- Optimize database queries
- Minimize re-renders

## Troubleshooting

### Common Issues

#### Backend won't start
- Check Rust version
- Verify dependencies
- Check port availability

#### Frontend build fails
- Clear node_modules
- Update dependencies
- Check Node.js version

#### Metabase connection issues
- Verify Metabase is running
- Check ports
- Test API endpoints

### Debugging Tips

```bash
# Check logs
tail -f logs/backend.log

# Test API endpoints
curl http://localhost:3000/api/v1/health

# Check database
sqlite3 data/datapulse.db "SELECT * FROM users;"
```

## Contributing

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature

# Commit changes
git commit -m "Add your feature"

# Push to remote
git push origin feature/your-feature

# Create pull request
```

### Code Reviews

- Follow existing patterns
- Write clear commit messages
- Include tests
- Document changes

## Resources

### Documentation

- [Rust Documentation](https://doc.rust-lang.org/)
- [Svelte Documentation](https://svelte.dev/docs)
- [Axum Documentation](https://docs.rs/axum)
- [Metabase Documentation](https://www.metabase.com/docs)

### Tools

- [Rust Analyzer](https://rust-analyzer.github.io/)
- [Svelte for VS Code](https://marketplace.visualstudio.com/items?itemName=svelte.svelte-vscode)
- [Docker](https://www.docker.com/)

## Support

For development questions, contact dev@datapulse.app