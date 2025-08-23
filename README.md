# Smart-Meeting-Scheduler
A simple Go-based meeting scheduler using the **Fiber** web framework. This API allows users to schedule meetings by finding common available time slots across participants’ calendars.
This service will accept a request to schedule a new meeting for a list of participants and will find the optimal time slot based on their existing calendars and a set of scheduling preferences.
## Table of Contents

- [Architecture & Design](#architecture--design)
- [Data Models](#data-models)
- [Algorithm & Heuristics](#algorithm--heuristics)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Postman Collection](#postman-collection)

---

## Architecture & Design

The application uses a lightweight in-memory datastore for simplicity. It is built using the **Fiber** web framework for fast HTTP handling.

- **Main Components**
  - `main.go` – Entry point of the application
  - `handlers` – API endpoints for scheduling meetings and fetching user calendars
  - `models` – Defines data structures (`User`, `Event`, `MeetingRequest`, `MeetingResponse`)
  - `utils` – Utility functions for date parsing and scoring

- **Design Choices**
  - **Fiber Framework**: Chosen for its speed and simplicity in building HTTP APIs in Go.
  - **In-memory DB**: Used slices for `users` and `events` to avoid external dependencies. This makes the app lightweight and easy to run locally.
  - **UUIDs**: Used for unique IDs for meetings and events.
  - **Time handling**: All times are in RFC3339 format for consistency and timezone support.
Steps

1.Clone the repository:
git clone https://github.com/your-username/meeting-scheduler.git
cd meeting-scheduler

2.Install dependencies:
go mod tidy

3.Run the application:
go run main.go

4.The API will be available at:
http://localhost:8080

5.API Endpoints
1. Schedule Meeting

POST /schedule-http://localhost:8080/schedule

Request Body:

{
  "participantIds": ["user1", "user2"],
  "durationMinutes": 30,
  "timeRange": {
    "start": "2025-08-01T09:00:00Z",
    "end": "2025-08-05T17:00:00Z"
  }
}

2. Get User Calendar

GET -(http://localhost:8080/users/user2/calendar?start=2025-08-01T00:00:00Z&end=2025-08-05T23:59:59Z)
