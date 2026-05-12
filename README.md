# SkillSprint

SkillSprint is a full-stack learning and skill assessment platform. It allows users to take quizzes, participate in live arenas, and track their training and test attempts.

## Project Structure

The project is divided into two main components:

- **`frontend/`**: A modern React application built with Next.js, featuring a robust UI using Tailwind CSS and Radix UI components. It also utilizes Prisma ORM for database interactions.
- **`go-backend/`**: A robust REST API backend built in Golang using the Gin framework. It handles core business logic, user authentication (including local and Google OAuth), and database management using GORM and SQLite.

## Core Features

- **User Authentication**: Secure signup and login functionality with role-based access control (User/Admin). Supports both local credentials and Google OAuth.
- **Arenas & Quizzes**: Participate in live "Arenas" (timed competitive quizzes) or take practice tests across various topics and difficulty levels.
- **Dynamic Questions**: Support for different question types including Multiple Choice Questions (MCQs) and subjective questions.
- **Progress Tracking**: Track user attempts, scores, and training progress over time.
- **Community Chat**: Real-time global chat room where users can send messages, share notes, images, and PDF files with each other. Features WebSocket-based instant messaging, file uploads, and persistent chat history.

## Tech Stack

### Frontend
- **Framework**: Next.js (React)
- **Styling**: Tailwind CSS & Radix UI
- **ORM**: Prisma
- **Form Handling**: React Hook Form & Zod
- **Icons**: Lucide React

### Backend
- **Language**: Golang
- **Framework**: Gin Web Framework
- **ORM**: GORM
- **Database**: SQLite (`dev.db`)
- **Authentication**: JWT (JSON Web Tokens) & bcrypt

## Getting Started

### Prerequisites
- Node.js (v18+)
- Go (1.25+)

### Running the Backend
1. Navigate to the `go-backend` directory.
2. Run `go mod tidy` to install dependencies.
3. Start the server: `go run main.go`

### Running the Frontend
1. Navigate to the `frontend` directory.
2. Install dependencies: `npm install`
3. Run the development server: `npm run dev`

## Project Timeline (Week-wise Description)

This section outlines the progressive development of the SkillSprint platform over a typical 6-week On-The-Job Training (OJT) period:

### Week 1: Requirement Analysis & Setup
- Understanding the core objectives of the SkillSprint platform.
- Initializing the project repositories and environments.
- Setting up the initial architecture for both the Next.js frontend and the Golang backend.
- Designing the database schema and establishing SQLite connections.

### Week 2: Backend Development & Authentication
- Implementing core models (User, Arena, Question, Attempt) using GORM.
- Developing robust REST APIs using the Gin framework.
- Setting up JWT (JSON Web Token) authentication and bcrypt password hashing.
- Integrating Google OAuth for seamless user login.

### Week 3: Frontend Development (UI/UX)
- Designing the user interface with Tailwind CSS and Radix UI components.
- Building responsive page layouts, dashboards, and navigation components.
- Implementing user registration and login forms with React Hook Form and Zod validation.

### Week 4: Frontend-Backend Integration
- Connecting the frontend Next.js application to the Golang backend APIs.
- Implementing state management, API data fetching, and secure route handling.
- Creating the user dashboard to accurately track training progress and test attempts.

### Week 5: Core Features Implementation (Arenas & Quizzes)
- Developing the logic for live Arenas and real-time quiz participation.
- Supporting multiple question formats (MCQs, subjective) in the frontend interface.
- Implementing backend scoring logic and leaderboards.

### Week 6: Testing, Refinement & Documentation
- Conducting comprehensive testing across the full stack to ensure reliability.
- Fixing bugs, optimizing database queries, and refining the user experience.
- Preparing project documentation (including this README) and preparing for final evaluation.

### Additional Features: Community Chat
- Implemented real-time WebSocket-based chat system for global community communication.
- Added support for text messages, formatted notes, image sharing (JPEG/PNG/GIF), and PDF file sharing.
- Integrated file upload with validation (10MB limit, type checking, filename sanitization).
- Built persistent chat history with database storage and auto-scroll functionality.
- Designed responsive chat UI matching the SkillSprint cyberpunk theme.

## Chat Feature Documentation

For detailed information about the Community Chat feature, see:
- **[CHAT_FEATURE.md](CHAT_FEATURE.md)** - Complete feature documentation
- **[QUICK_START_CHAT.md](QUICK_START_CHAT.md)** - Quick start guide
- **[CHAT_ARCHITECTURE.md](CHAT_ARCHITECTURE.md)** - Architecture and data flow
- **[CHAT_TESTING_GUIDE.md](CHAT_TESTING_GUIDE.md)** - Testing checklist
