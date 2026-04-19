# 🏋️ OpenWorkout

An open-source, extensible workout tracking platform designed for flexibility, intelligence, and community-driven innovation.

> ⚠️ **Status: Work in Progress (WIP)**  
> This project is actively being developed. Features, APIs, and architecture may change frequently.

---

## ✨ Vision

The goal of this project is to build a modern workout ecosystem that:

- Tracks workouts with precision and simplicity
- Learns from user behavior to provide intelligent recommendations
- Supports plugins for community-driven extensions
- Works seamlessly across mobile and web

---

## 🧱 Architecture Overview

This project uses a **monorepo structure** with multiple apps and services:
```
/apps
/mobile → React Native app
/web → Web app (React)
/landing → Marketing site

/services
/workout → Workout logging service (Go)
/exercise → Exercise recommendation service (Go)
/ml → Machine learning & embeddings service (Go)

/packages
/shared → Shared types, utilities
/ui → Shared UI components

/docs → Documentation
```


---

## ⚙️ Tech Stack

### Frontend
- React Native (mobile)
- React (web)
- TypeScript

### Backend
- Go (microservices architecture)
- REST / gRPC (TBD)

### Data & ML
- PostgreSQL
- Vector embeddings (user + workout representation)
- Nearest-neighbor search (planned)

---

## 🧠 Core Concepts

### 1. Workout Tracking (WIP)
Users can:
- Log exercises, sets, reps, weights
- Track progression over time

### 2. User Embeddings (WIP)
Each user is represented as a vector based on:
- Profile data (age, goals, etc.)
- Workout behavior (exercise preferences, volume, intensity)

These embeddings are updated after each workout.

### 3. Smart Recommendations (Planned)
- Suggest exercises based on similar users
- Replace exercises dynamically when needed
- Adapt to user goals and training style

### 4. Plugin System (Planned)
Developers will be able to:
- Extend functionality via plugins
- Add new recommendation strategies
- Integrate external services

---

## 🚧 Current Progress

- [ ] Initial project structure
- [ ] Workout service (basic scaffolding)
- [ ] Exercise service
- [ ] ML service (user embeddings)
- [ ] Mobile app MVP
- [ ] Web app MVP
- [ ] Plugin system

---

## 🚀 Getting Started

### Prerequisites

- Node.js (>= 18)
- Go (>= 1.21)
- Docker (optional, recommended)

---
