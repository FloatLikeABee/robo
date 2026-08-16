# Academi

An AI-powered academic ecosystem for students and knowledge seekers.

## Project Setup

This project is a React Native application based on the design specifications in `design.md`.

### Directory Structure

```
├── assets/          # Static assets (images, icons, etc.)
├── components/      # Reusable UI components
├── screens/         # Application screens
├── navigation/      # Navigation configuration
├── store/           # State management (Zustand)
├── services/        # API services and business logic
├── themes/          # Theme configuration
├── utils/           # Utility functions
├── constants/       # Constant values
├── App.js           # Main application component
├── index.js         # Entry point
├── package.json     # Project dependencies and scripts
└── ...              # Configuration files
```

### Auth (UsersPanel / Morph)

Academi uses the same UsersPanel login as Morph. From MorphAI’s apps menu, opening Academi passes `?userspanel_token=…`, which the web app exchanges via `POST /api/v1/auth/dev-login`.

Set in `backend/.env`:

```bash
USERS_PANEL_BASE_URL=http://127.0.0.1:5001
```

Or from the workspace root: `./start-all.sh start academi` (API `:8978`, web `:8765`).

### Installation

1. Install dependencies:
   ```
   npm install
   ```

2. Run the application:
   ```
   # API + web together
   ./start.sh

   # For Android
   npm run android

   # For iOS
   npm run ios

   # For Web (HTML5 smartphone browser version)
   # Simply open web/index.html in your browser
   # Or serve it with a local server for better experience
   ```

### Design System

The application follows the design system outlined in `design.md`:
- Dark mode default with `#0A0F1C` background
- Gradient colors: `#5B8CFF` → `#9B6DFF` → `#00D4FF`
- Glassmorphism effects with blurred surfaces
- Rounded corners (2xl+)
- Micro-interactions and animations
- Futuristic monospace icons inspired by synthwave aesthetics

### Components Implemented

- Button (primary, secondary, ghost, icon variants)
- ChatBubble (user and AI variants with source preview)
- Basic screen templates for all main sections

### Web Version

A responsive HTML5 web version is available in the `web/` directory, optimized for smartphone browsers:

- **Features**: All main screens (Chat, Community, Docs, Profile) with full interactivity; chat **Learning guide** mode (web research + Markdown summary)
- **Design**: Matches the mobile app with dark theme, glassmorphism, and responsive layout
- **Browser Support**: Modern mobile browsers with touch interactions
- **How to run**: Serve the `web/` folder with a local HTTP server (required for ES modules), then open `index.html`. Example: `python3 -m http.server 8765` from `web/`, then visit `http://localhost:8765`.
- **Source layout**: `web/js/` — `api.js`, `chat.js`, `docs.js`, `community.js`, `app.js`, `main.js` (plus `markdown.js` for shared rendering).

### Next Steps

1. Implement actual functionality for each screen
2. Add state management with Zustand
3. Integrate AI services
4. Implement community features
5. Add documentation and guide modules
6. Implement theming system with light/dark toggle