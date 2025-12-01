# Cycle Scene

A platform-agnostic application that helps people discover bike rides in their city. Cycle Scene aggregates bike events from community organizers like Shift2Bikes and allows community members to submit their own rides, making bike events accessible across all platforms and devices.

## Quick Links

- **Main Directory**: https://cyclescene.cc
- **Portland PWA**: https://pdx.cyclescene.cc
- **GitHub**: https://github.com/spacesedan/cyclescene

## Project Overview

Cycle Scene consists of multiple interconnected services:

### Frontend Applications
- **Directory** - City-based directory for selecting which PWA to access
- **PWA** - Mobile-first Progressive Web App with map view, offline support, and ride discovery
- **Form** - User interface for submitting new bike rides
- **Dashboard** - Admin analytics and ride management

### Backend Services
- **API** - REST API for ride management, submissions, and authentication
- **Scraper v2** - Automated scraper for Shift2Bikes events with geocoding
- **Image Optimizer** - Image processing and optimization service
- **Token Cleaner** - Automated job to clean up expired submission tokens
- **DB Backups** - Automated database backup service

### Infrastructure
- **Google Cloud Platform** - Cloud Run, Cloud Storage, Cloud Scheduler
- **TursoDB** - SQLite-compatible serverless database
- **Terraform/OpenTofu** - Infrastructure as Code

## Architecture

```
Frontend Applications (Svelte/SvelteKit)
├── Directory
├── PWA
├── Form
└── Dashboard
        |
        v
┌─────────────────────┐
│   REST API          │
│  (Cloud Run)        │
└──────────┬──────────┘
           |
           v
┌─────────────────────┐
│  TursoDB (SQLite)   │
│  - Rides & Events   │
│  - Groups           │
│  - Submissions      │
│  - Geocache         │
│  - API Keys         │
└──────────┬──────────┘
           |
           v
Background Jobs
├── Scraper v2
├── Token Cleaner
├── Image Optimizer
└── DB Backups
```

## Getting Started

### Prerequisites
- Go 1.24+
- Node.js 18+ (for frontends)
- Docker
- GCP Account (for cloud deployment)

### Local Development

#### Backend Setup
```bash
cd functions
go mod download
# See /functions/README.md for detailed setup
```

#### Frontend Setup
```bash
cd frontends/{form|directory|pwa|dashboard}
npm install
npm run dev
```

### Deployment

```bash
# Bootstrap GCP infrastructure
make bootstrap-infra

# Deploy all services
make deploy-all
```

See individual service READMEs for specific deployment instructions.

## Directory Structure

```
cycle-scene/
├── functions/              # Go backend services
│   ├── cmd/
│   │   ├── api/           # REST API
│   │   ├── scraperv2/     # Event scraper
│   │   ├── image-optimizer/
│   │   ├── token-cleaner/
│   │   └── db-backups/
│   ├── internal/          # Shared packages
│   └── README.md
├── frontends/             # SvelteKit applications
│   ├── directory/         # City directory
│   ├── pwa/              # Mobile app
│   ├── form/             # Ride submission
│   ├── dashboard/        # Admin dashboard
│   └── README.md
├── db/                    # Database migrations
├── infrastructure/        # Terraform modules
└── README.md
```

## Technology Stack

**Backend:**
- Go 1.24
- TursoDB (SQLite-compatible)
- Chi HTTP router
- Google Cloud Platform

**Frontend:**
- Svelte 5
- SvelteKit 2
- Vite 7
- TailwindCSS 4
- MapLibre GL (PWA)

**Infrastructure:**
- OpenTofu/Terraform
- Docker
- Google Cloud Run
- Cloud Scheduler

## Data Flow

1. **Ride Discovery**: Users browse rides in PWA or Form frontend
2. **Ride Submission**: Users submit new rides via Form frontend
3. **Route Visualization**: Users can view route maps and elevation profiles for rides with associated routes
4. **Data Aggregation**: Scraper v2 fetches Shift2Bikes events every 3 hours
5. **Image Processing**: Uploaded images are optimized and cached
6. **Analytics**: Directory app tracks usage patterns
7. **Maintenance**: Token Cleaner and DB Backups run on schedules

## Privacy and Analytics

- **Non-invasive analytics**: Only visit tracking and CTA clicks
- **No user tracking**: No IP addresses, cookies, or personal data stored
- **No login required**: Anonymous, ephemeral usage
- **Transparent**: Public privacy policy with full disclosure

See `/frontends/directory/src/routes/privacy/+page.svelte` for privacy policy details.

## Contributing

We welcome contributions! Please see individual service READMEs for development guidelines.

## License

See LICENSE file for details.

## Support

For issues, feature requests, or data usage inquiries, please open an issue on GitHub.

---

Built with care for the bike community
