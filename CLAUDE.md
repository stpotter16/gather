# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**Gather** — a family gathering coordination website. Users create events, invite family members, and collaborate on logistics: arrival/departure itineraries, food planning, and activity brainstorming.

## Planned stack

- **Backend:** SSR Go + PostgreSQL (Neon)
- **Hosting:** Fly.io
- **Styling:** Tailwind CSS

## Auth and user management

There is no self-serve signup. User rows are added to the database manually by the admin. The invite flow is restricted to users already in the system — inviting someone who doesn't have an account is not supported.

## Screens

HTML mockups live in `screens/`. The main event page (`screens/event.html`) is the reference for the design system — color palette, component patterns, and layout conventions all derive from it.

- `screens/gatherings.html` — home screen, lists upcoming and past events
- `screens/create-gathering.html` — form to create a new event
- `screens/event.html` — full event detail page (tabbed: Overview, Itineraries, Food, Activities); includes an RSVP banner for users who haven't responded yet
- `screens/invite.html` — modal for inviting existing users to an event
- `screens/add-itinerary.html` — modal for adding arrival/departure details; supports Flying (with flight number lookup), Driving, and Other
- `screens/event-header.html` — header component in isolation
- `screens/event-attendees.html` — attendees section in isolation
- `screens/event-itineraries.html` — itineraries section in isolation
- `screens/event-food.html` — food coordination section in isolation
- `screens/event-activities.html` — activity brainstorm section in isolation

## Third-party integrations

**Flights:** flight number lookup via a flight data API (e.g. AviationStack) — user enters a flight number and route/times auto-fill. Triggered from the add-itinerary modal.

**Accommodations:** no API integration possible (Airbnb and VRBO have no public API). Accommodations are stored as labelled links. The accommodations list lives on the Overview tab and is flexible — there can be one shared rental or many. Anyone in the event can add a link.

## Design conventions

- **Accent color:** amber (`bg-amber-500`, `text-amber-600`)
- **Neutral base:** stone scale (`bg-stone-100` page background, `bg-white` cards)
- **Cards:** `bg-white rounded-xl border border-stone-200`
- **Section headers inside cards:** `text-xs font-semibold text-stone-400 uppercase tracking-wide`
- **Avatars:** colored `rounded-full` divs with a single capital letter initial; each person has a consistent color across all screens
- **Tailwind via CDN** for now (`<script src="https://cdn.tailwindcss.com"></script>`)
