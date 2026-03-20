# EventPass Backend Modules - Development Roadmap

**Total: 9 Core Modules** | **Build Order: 1→9** | **Estimated: 2-3 weeks solo**

## 🎯 Module 1: Database & Migrations
**Purpose**: Initialize PostgreSQL connection + core table schemas
**Files**: `internal/database/`
- GORM connection with connection pooling
- Auto-migrations for: `users`, `events`, `attendees`, `forms`, `checkins`
- Seed initial admin user
**Endpoints**: Health check `/health`
**Status**: Foundation - Build FIRST

## 🔐 Module 2: Auth (Organizer-Only)
**Purpose**: JWT authentication for event organizers only
**Files**: `internal/middleware/`
- POST `/api/register` - organizer signup
- POST `/api/login` - JWT token generation  
- Middleware: `AuthRequired()` - protect all organizer routes
- Rate limiting on auth endpoints
**Endpoints**: `/api/auth/*`
**Status**: Security gate - Build SECOND

## 📋 Module 3: Events CRUD
**Purpose**: Organizers create/list/manage events
**Files**: `internal/api/events.go`
- POST `/api/events` - create event (title, date, capacity, public/private)
- GET `/api/events` - list organizer's events
- GET `/api/events/:id` - event details
- PATCH `/api/events/:id` - update capacity/status
- DELETE `/api/events/:id`
**DB Tables**: `events` (id, organizer_id, title, date, capacity, is_public, status)
**Status**: Core business object

## 📝 Module 4: Custom Forms Builder
**Purpose**: Dynamic form fields for each event
**Files**: `internal/service/forms.go`
- POST `/api/events/:id/forms` - add custom fields (text, dropdown, checkbox)
- GET `/api/events/:id/forms` - get form schema
- Form validation engine (JSON schema → GORM dynamic fields)
**DB Tables**: `form_fields` (event_id, field_name, field_type, required, options)
**Status**: Event-specific registration

## 🎟️ Module 5: Attendee Registration
**Purpose**: Public form submission → attendee record
**Files**: `internal/api/registration.go`
- POST `/api/events/:id/register` - public endpoint (no auth required)
- Validate dynamic form fields against event schema
- Generate attendee record + secure hash
- Capacity check: reject if full
**DB Tables**: `attendees` (id, event_id, form_data JSONB, hash, status, created_at)
**Status**: Public money-maker

## 💳 Module 6: Payments (Hyperswitch)
**Purpose**: Optional paid tickets via Hyperswitch
**Files**: `internal/service/payments.go`
- POST `/api/events/:id/pay` - create Hyperswitch session
- POST `/api/webhooks/hyperswitch` - verify payment success
- Update attendee status: `pending → paid`
**Endpoints**: Payment initiation + webhook
**Status**: Revenue module

## 🔍 Module 7: QR Check-in Scanner
**Purpose**: Secure QR validation at venue
**Files**: `internal/service/qr.go`, `internal/api/checkin.go`
- POST `/api/checkins/scan` - decode QR → verify hash → mark checked-in
- Atomic DB update (prevent duplicates)
- Capacity enforcement (stop at 100%)
**QR Format**: `eventpass://v1/{event_id}/{attendee_hash}`
**DB Tables**: `checkins` (attendee_id, scanned_at, volunteer_id)
**Status**: Venue gatekeeper

## 📊 Module 8: Real-time Analytics
**Purpose**: Live dashboard for organizers
**Files**: `internal/service/analytics.go`
- GET `/api/events/:id/analytics` - sales, checkins, capacity %
- Centrifugo integration for live updates
- CSV export endpoint
**Endpoints**: Dashboard data + WebSocket events
**Status**: Organizer delight

## 📧 Module 9: Notifications & Exports
**Purpose**: Post-registration emails + data portability
**Files**: `internal/service/notifications.go`
- Email attendee QR passes (HTML + PDF attachment)
- POST `/api/events/:id/notify` - bulk SMS/email
- GET `/api/events/:id/export` - CSV download (name, email, status)
**Status**: Final polish

---

## 🚀 Build Priority Order
