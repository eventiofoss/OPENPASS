# Eventio API Documentation

## Base URL
All API paths listed below are relative to the configured server host and port (default: `http://localhost:8080`).

## Authentication & Headers
For all protected routes (all routes under `/api/auth/me`, `/api/auth/logout`, and `/api/events/`), the API expects authentication via a session cookie:
- **Expected Headers:** `Cookie: session=<token>`

For endpoints accepting a JSON payload:
- **Expected Headers:** `Content-Type: application/json`

---

## Health Check

### GET `/api/health`
Check if the API and database are running successfully.

- **Expected Headers:** None
- **Expected request body:** None
- **Expected Output (200 OK):**
  ```json
  {
    "status": "ok",
    "message": "Database connection successful"
  }
  ```

---

## Authentication

### POST `/api/auth/register`
Register a new organizer account.

- **Expected Headers:** `Content-Type: application/json`
- **Expected Request Body:**
  ```json
  {
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "securepassword123"
  }
  ```
- **Expected Output (201 Created):**
  ```json
  {
    "message": "Account created successfully",
    "organizer": {
      "id": "uuid",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  }
  ```

### POST `/api/auth/login`
Authenticate an organizer and issue a session cookie.

- **Expected Headers:** `Content-Type: application/json`
- **Expected Request Body:**
  ```json
  {
    "email": "jane@example.com",
    "password": "securepassword123"
  }
  ```
- **Expected Output (200 OK):**
  - **Set-Cookie:** `session=<token>; HttpOnly; SameSite=Strict`
  ```json
  {
    "message": "Logged in successfully"
  }
  ```

### GET `/api/auth/me`
Retrieve the currently authenticated organizer.

- **Expected Headers:** `Cookie: session=<token>`
- **Expected Output (200 OK):**
  ```json
  {
    "organizer": {
      "id": "uuid",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  }
  ```

### POST `/api/auth/logout`
Clear the organizer session cookie.

- **Expected Headers:** `Cookie: session=<token>`
- **Expected Output (200 OK):**
  - **Set-Cookie:** `session=; HttpOnly; SameSite=Strict; Expires=past_date`
  ```json
  {
    "message": "Logged out successfully"
  }
  ```

---

## Events

All routes in this group require the `Cookie: session=<token>` header.

### POST `/api/events`
Create a new event owned by the authenticated organizer.

- **Expected Headers:** `Content-Type: application/json`, `Cookie: session=<token>`
- **Expected Request Body:**
  ```json
  {
    "title": "Tech Conference 2026",
    "description": "Annual tech meetup",
    "start_date": "2026-05-01T09:00:00Z",
    "venue": "Convention Center",
    "capacity": 500,
    "is_public": true
  }
  ```
- **Expected Output (201 Created):**
  ```json
  {
    "message": "Event created successfully",
    "event": {
      "id": "uuid",
      "slug": "tech-conference-2026"
    }
  }
  ```

### GET `/api/events`
List all events owned by the authenticated organizer.

- **Expected Headers:** `Cookie: session=<token>`
- **Expected Output (200 OK):**
  ```json
  {
    "events": [
      {
        "id": "uuid",
        "title": "...",
        "description": "...",
        "start_date": "...",
        "venue": "...",
        "capacity": 500,
        "is_public": true,
        "status": "active"
      }
    ]
  }
  ```

### GET `/api/events/:id`
Get details of a single organizer-owned event by its ID.

- **Expected Headers:** `Cookie: session=<token>`
- **Expected Output (200 OK):**
  ```json
  {
    "event": {
      "id": "uuid",
      "title": "...",
      "...": "..."
    }
  }
  ```

### PATCH `/api/events/:id`
Partially update an organizer-owned event.

- **Expected Headers:** `Content-Type: application/json`, `Cookie: session=<token>`
- **Expected Request Body:**
  ```json
  {
    "capacity": 600,
    "status": "full"
  }
  ```
- **Expected Output (200 OK):**
  ```json
  {
    "message": "Event updated successfully",
    "event": {
      "id": "uuid",
      "title": "...",
      "capacity": 600,
      "status": "full"
    }
  }
  ```

### DELETE `/api/events/:id`
Delete an organizer-owned event.

- **Expected Headers:** `Cookie: session=<token>`
- **Expected Output (200 OK):**
  ```json
  {
    "message": "Event deleted successfully"
  }
  ```

---

## Forms

All routes in this group require the `Cookie: session=<token>` header.

### POST `/api/events/:id/forms`
Replaces all existing form fields for a given event.

- **Expected Headers:** `Content-Type: application/json`, `Cookie: session=<token>`
- **Expected Request Body:**
  ```json
  {
    "fields": [
      {
        "name": "tshirt_size",
        "type": "select",
        "label": "T-Shirt Size",
        "required": true,
        "position": 1,
        "options": ["S", "M", "L", "XL"]
      }
    ]
  }
  ```
- **Expected Output (201 Created):**
  ```json
  {
    "message": "Form fields saved successfully",
    "fields": [
      {
        "name": "tshirt_size",
        "type": "select",
        "label": "T-Shirt Size",
        "required": true,
        "position": 1,
        "options": ["S", "M", "L", "XL"]
      }
    ]
  }
  ```

### GET `/api/events/:id/forms`
Retrieve the form schema/fields configured for an event.

- **Expected Headers:** `Cookie: session=<token>`
- **Expected Output (200 OK):**
  ```json
  {
    "fields": [
      {
        "name": "tshirt_size",
        "type": "select",
        "label": "T-Shirt Size",
        "required": true,
        "position": 1,
        "options": ["S", "M", "L", "XL"]
      }
    ]
  }
  ```
