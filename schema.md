# Eventio Storage Design

Your current direction is strong, but it was not fully "perfect" for this flow yet.

The biggest fixes were:

1. `slug` should not be the secret URL for private events. Human-readable slugs are guessable, so private events need a separate cryptographic `access_token`.
2. `attendees.email` should be unique per event, not just indexed globally, so the same person cannot register twice for the same event by accident.
3. `public_url` for attendee passes should be derived at runtime from host + route + `qr_hash`, not stored in the database.
4. Organizer-side accounts should be named consistently as `organizers`, since attendees are passwordless.
5. `updated_at` fields are useful on mutable tables so organizer changes are auditable.

Below is the revised schema that matches the current Go models.

## 1. `organizers`
Authenticated backend users only.

```text
- id: UUID PK
- name: string
- email: string UNIQUE
- password_hash: string
- role: enum('organizer','admin','volunteer')
- created_at: timestamp
- updated_at: timestamp
```

Stores: organizer accounts, admin access, volunteer scanner access.

## 2. `events`
Master event record + discovery controls.

```text
- id: UUID PK
- slug: string UNIQUE
- access_token: string UNIQUE
- organizer_id: UUID FK -> organizers.id
- title: string
- description: text
- start_date: timestamp
- venue: string
- capacity: integer
- is_public: boolean DEFAULT false
- status: enum('draft','active','full','cancelled')
- total_registered: integer DEFAULT 0
- created_at: timestamp
- updated_at: timestamp
```

Notes:
- `slug` is the public-friendly route.
- `access_token` is the secret/private route token.
- `total_registered` is a cached counter; the hard capacity rule should still be enforced inside a transaction when creating registrations.

## 3. `form_fields`
Organizer-defined registration form fields.

```text
- id: UUID PK
- event_id: UUID FK -> events.id
- name: string
- type: enum('text','email','select','checkbox','number')
- label: string
- required: boolean
- options: JSONB NULL
- position: integer
- created_at: timestamp
- updated_at: timestamp
```

Recommended constraints:
- UNIQUE (`event_id`, `name`)
- UNIQUE (`event_id`, `position`)

## 4. `attendees`
Passwordless registrations for an event.

```text
- id: UUID PK
- event_id: UUID FK -> events.id
- email: string
- name: string
- form_data: JSONB
- qr_hash: string UNIQUE
- status: enum('registered','paid','checked_in','cancelled') DEFAULT 'registered'
- checked_in_at: timestamp NULL
- created_at: timestamp
- updated_at: timestamp
```

Recommended constraints:
- UNIQUE (`event_id`, `email`)

Notes:
- Do not store `public_url`; derive it from the deployment base URL plus `qr_hash`.
- `qr_hash` should be generated with a cryptographically secure random token.

## 5. `payments`
Current payment record for a registration.

```text
- id: UUID PK
- attendee_id: UUID FK -> attendees.id UNIQUE
- hyperswitch_payment_id: string UNIQUE
- amount: decimal(10,2)
- currency: char(3)
- status: enum('pending','succeeded','failed')
- paid_at: timestamp NULL
- created_at: timestamp
- updated_at: timestamp
```

Note:
- This 1:1 version is fine for an MVP. If you later want full retry/audit history, split this into payment attempts.

## 6. `checkins`
Successful gate scans.

```text
- id: UUID PK
- attendee_id: UUID FK -> attendees.id UNIQUE
- scanned_by: UUID FK -> organizers.id
- scanned_at: timestamp
- created_at: timestamp
```

Notes:
- This schema stores the first successful entry.
- If you later want every scan attempt logged, add a separate scan-attempt audit table.

## 7. `exports`
Audit log for attendee data exports.

```text
- id: UUID PK
- event_id: UUID FK -> events.id
- organizer_id: UUID FK -> organizers.id
- format: enum('csv','excel')
- filename: string
- generated_at: timestamp
```

## Relationships

```text
organizers 1:M events
events 1:M form_fields
events 1:M attendees
attendees 1:1 payments
attendees 1:1 checkins
events 1:M exports
organizers 1:M exports
organizers 1:M checkins
```

## Implementation Notes

```text
✅ JSONB handles custom form payloads cleanly
✅ access_token protects private event links
✅ qr_hash secures attendee pass links
✅ UNIQUE(event_id, email) prevents duplicate registrations
✅ total_registered works as a fast counter
✅ capacity enforcement should happen inside the registration transaction
✅ updated_at on mutable tables helps admin auditing
```

## GORM Migration Order

```text
1. organizers
2. events
3. form_fields
4. attendees
5. payments
6. checkins
7. exports
```
