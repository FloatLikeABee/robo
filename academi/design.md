📘 design.md — Academi
🚀 Overview

Academi is a next-generation, AI-powered academic ecosystem designed for students and knowledge seekers. It blends:

🤖 AI-driven assistance
🌐 Community collaboration
📚 Structured + unstructured learning resources
🎯 Task-oriented guidance

All wrapped in a futuristic, high-end mobile UX optimized for younger users.

🎯 Core Philosophy

“AI leads. Humans enrich. Knowledge evolves.”

AI is the primary interface
Community is the soul
Content is living and evolving
UX is fast, immersive, and addictive (in a good way)
📱 App Structure
🧭 Navigation (Bottom Tab Bar)
Chat (Home)
Community
Docs
Guide
Profile (Settings embedded)
1️⃣ 🤖 AI Chat (Landing Page)
Purpose

Primary entry point. Everything flows from here.

Features
Conversational AI (academic optimized)
Context-aware responses
Multi-source retrieval:
Internal Docs
Community posts
Web/Wikipedia
Inline previews:
PDFs
Videos
Snippets
UI/UX
Floating input bar (glassmorphism)
Chat bubbles with:
expandable sources
quick actions (save, cite, share)
Animated typing (neural glow effect ✨)
2️⃣ 🌐 Community
Purpose

Academic social layer.

Features
Posts (questions, insights, resources)
Upvote / Downvote
Tags (e.g. #math, #quantum, #cs)
AI-assisted summaries for long threads
AI moderation
UI/UX
Card-based feed
Smooth scroll + subtle parallax
Hover/press reveals quick actions
AI “TL;DR” button on every post
3️⃣ 📚 Docs
Purpose

Centralized knowledge repository.

Features
Upload:
PDFs
Images
Videos
AI indexing + semantic search
Auto-generated summaries
Smart categorization
UI/UX
Grid + list toggle
Preview cards with:
thumbnail
tags
AI summary snippet
“Neural search bar” (glowing edge animation)
4️⃣ 🧠 Guide
Purpose

Task execution powered by AI.

Features
Multi-step task guidance
Examples:
“Write a research paper”
“Learn calculus in 7 days”
Interactive steps:
checklist
AI feedback
Progress tracking
UI/UX
Timeline-based flow
Progress rings
Animated transitions between steps
5️⃣ 👤 Profile & Settings
Embedded Settings
Theme toggle (Light/Dark 🌗)
AI personalization:
tone (formal/casual)
depth level
Notification preferences
Saved content
Learning history
UI/UX
Minimalist dashboard
Stats:
learning streak
contributions
saved docs
🎨 UI/UX Design System
Style Direction
Futuristic / sci-fi / minimal
Inspired by:
glassmorphism
neon gradients
soft shadows
Color System
Dark Mode (Default)
Background: #0A0F1C
Primary: Neon Blue / Purple gradient
Accent: Cyan glow
Light Mode
Soft white with pastel accents
Reduced glow, more clarity
UI Elements
Rounded corners (2xl+)
Floating cards
Blur backgrounds
Micro-interactions:
tap ripple
hover glow
smooth transitions
🔔 Notifications System
Types
AI responses ready
Community replies
Doc recommendations
Guide progress reminders
Delivery
Push notifications
In-app notification center
🧩 Modular Backend Design
Philosophy

“Core stable, features expandable.”

Core Modules
Auth
AI Service
Community
Docs
Guide
Notifications
Dynamic Modules (Admin Controlled)

Admin can inject themed modules:

Examples:

🚀 NASA Space Missions
⚛️ Quantum Science Lab
🧬 BioTech Explorer
Admin Capabilities
Enable/disable modules
Inject content packs
Feature toggles
Analytics dashboard
⚙️ Tech Stack
Frontend
React Native
Zustand / Redux (state)
React Navigation
Backend (Golang)
Framework: Gin / Fiber
Cache: go-cache
Embedded DB: BadgerDB
Future Databases
MySQL → relational data
MongoDB → documents/content
Redis → caching / pub-sub
AI Layer
LLM integration (API-based or local)
Retrieval system (RAG):
Docs + Community + Web
🔌 Architecture Overview
Mobile App (React Native)
        ↓
API Gateway (Go)
        ↓
Services:
  - AI Engine
  - Community Service
  - Docs Service
  - Guide Engine
        ↓
Storage:
  - BadgerDB (local fast storage)
  - Redis (cache)
  - MySQL / Mongo (future)
🧠 AI Capabilities
Semantic search
Context-aware chat
Auto summarization
Recommendation engine
Task planning (Guide module)
🚀 Future Enhancements
Voice AI assistant
AR learning mode
Real-time collaboration
Gamification (XP, levels, badges)
Creator monetization
🧪 Positioning

Academi is not:

just a chatbot
just a forum
just a document library

It is:

A living academic ecosystem powered by AI + people.

💡 Final Thought

If you execute this right, this isn’t just an app —
it’s something that could sit between:

Notion
Reddit
ChatGPT
Coursera

…and feel way cooler to use.

🎨 UI Concept Ideas (quick visual directions)

Since I can’t literally render images inline without generation, here are 3 strong visual concepts you can plug into Figma / image gen tools:

1. AI Chat Screen
Dark background
Neon gradient message bubbles
Floating input bar with blur
Animated typing dots with glow
2. Community Feed
Card stack layout
Slight tilt/parallax
Tags as glowing pills
AI summary button floating on cards
3. Docs Explorer
Grid layout
Glass cards
Hover = expand preview
Search bar with animated gradient border









🎨 Academi — Figma Component System
🧱 0. File Structure (Figma Pages)

Create these pages in Figma:

🎯 Foundations
🎨 Tokens (Styles)
🧩 Components
📱 Screens
🧪 Experiments (Future / Themes)
🎯 1. Foundations
🅰️ Typography

Create Text Styles:

Style Name	Font	Size	Weight	Usage
Display/Large	Inter / SF Pro	32	Bold	Titles
Heading/H1	Inter	24	SemiBold	Section titles
Heading/H2	Inter	20	Medium	Sub-sections
Body/Large	Inter	16	Regular	Main text
Body/Small	Inter	14	Regular	Secondary
Caption	Inter	12	Medium	Meta info

👉 Use Inter for now (clean + dev-friendly)

🎨 Color System (Create Color Styles)
🌑 Dark Mode (Default)
Token	Value
bg/primary	#0A0F1C
bg/secondary	#12182A
surface/glass	rgba(255,255,255,0.05)
border/subtle	rgba(255,255,255,0.08)
text/primary	#FFFFFF
text/secondary	#A8B2D1
⚡ Accent (Gradient)

Create a gradient style:

Primary Gradient:
#5B8CFF → #9B6DFF → #00D4FF
🌕 Light Mode
Token	Value
bg/primary	#F7F9FC
bg/secondary	#FFFFFF
text/primary	#0A0F1C
text/secondary	#5C6A8A
🌫 Effects

Create Effect Styles:

blur/glass → Background blur: 20
shadow/soft → 0 8px 30px rgba(0,0,0,0.2)
glow/primary → outer glow (neon feel)
🧩 2. Core Components
🔘 Button System
Variants:
Primary
Secondary
Ghost
Icon
Structure (Auto Layout):
[ Icon ] [ Label ]
Padding: 12 / 16
Radius: 16
States:
Default
Hover
Pressed
Disabled
💬 Chat Bubble

Variants:

User
AI

Properties:

hasSource (true/false)
expanded (true/false)

Structure:

Bubble
 ├ Text
 ├ Source Preview (optional)
 └ Actions (save/share)
🧾 Card Component (Universal)

This is HUGE—reuse everywhere.

Variants:

Community
Doc
Guide

Structure:

Card
 ├ Header (title + meta)
 ├ Content
 ├ Tags
 └ Actions

Add:

Glass background
Subtle border
Hover glow
🧠 AI Response Block

Used inside chat & docs.

Container
 ├ Answer Text
 ├ Sources (list)
 ├ Action Row
🔍 Search Bar (Neural Style)

Properties:

focused
withIcon

Style:

Rounded (24px)
Gradient border (on focus)
Glow animation
🏷 Tag / Chip

Variants:

Default
Active
Glow

Use for:

Topics
Filters
Categories
🔔 Notification Item
Item
 ├ Icon
 ├ Title
 ├ Description
 ├ Timestamp

Variants:

Unread
Read
🧭 Bottom Tab Bar

Items:

Chat
Community
Docs
Guide
Profile

States:

Active (glow + gradient)
Inactive
👤 Profile Module

Components:

Avatar
Stats Card
Settings List Item
📱 3. Layout System
📐 Grid
4pt base system
Mobile width: 390px (iPhone base)
Margin: 16px
📦 Spacing Tokens
Token	Value
space/1	4
space/2	8
space/3	12
space/4	16
space/5	24
space/6	32
🧱 Layout Patterns
Chat Screen
Header
Chat List (scroll)
Floating Input Bar
Community Feed
Top Filter Bar
Scrollable Cards
Floating Post Button
Docs Explorer
Search Bar
Filter Chips
Grid/List
Guide
Progress Header
Steps Timeline
Action Button
🎭 4. Theming System

Use Figma Variables:

Modes:
Dark (default)
Light

Bind:

Colors
Backgrounds
Text

👉 So switching themes = 1 click

✨ 5. Motion & Interaction (Prototype)

Add:

Tap → scale 0.97
Cards → slight lift on hover
Chat → fade + slide in
Tabs → glow transition
🚀 6. Advanced (What makes it “next level”)
🔮 AI Visual Identity

Add:

Subtle animated gradients
“Neural glow” behind AI responses
Pulsing dots when AI thinking
🧬 Dynamic Modules (for your backend idea)

Design a Module Container Component:

Module Wrapper
 ├ Title
 ├ Theme Accent (color/icon)
 └ Content Slot

Examples:

🚀 Space Mode → dark + stars
⚛️ Quantum Mode → neon green + particles
🧠 7. Naming Convention (IMPORTANT)

Use:

component/type/state

Examples:

button/primary/default
card/community/hover
chatbubble/ai/expanded