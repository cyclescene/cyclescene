# Dashboard

Admin dashboard for viewing analytics and managing rides.

## Overview

The Dashboard provides admin access to:
- Analytics and metrics
- Ride management
- User submissions
- Performance monitoring
- Content moderation

**Technology**: SvelteKit, Svelte 5, TailwindCSS
**URL**: https://dashboard.cyclescene.cc
**Deployment**: Vercel
**Access**: Admin API key required

## Getting Started

```bash
cd frontends/dashboard
npm install
npm run dev
```

Access at `http://localhost:5173`

## Dashboard Views

### Analytics
- Total rides and submissions
- Traffic trends
- City-level statistics
- Source attribution data

### Ride Management
- View all submitted rides
- Approve/reject submissions
- Edit ride details
- Manage ride images

### User Submissions
- Queue of pending rides
- Flagged content
- Submission timestamps
- Submitter information

### Performance
- Page load metrics
- API response times
- Error tracking
- System health

## Authentication

Uses admin API key for access. Admins must provide their API key to authenticate with the dashboard.

Generate admin keys using the admin-keys service:
```bash
cd functions/cmd/admin-keys
go run main.go generate
```

## Data Updates

Dashboard pulls data from:
- Rides API (authenticated with admin key)
- Analytics database
- User submissions queue

## Styling

- TailwindCSS for styling
- Dark mode support
- Responsive dashboard layout
- Charts and data visualizations

## Troubleshooting

### Dashboard Not Loading Data
- Verify admin API key is valid
- Check API connectivity
- Review browser console for errors

### Slow Analytics Loading
- Check API response times
- Consider implementing caching
- Optimize database queries

### Auth Issues
- Verify admin API key is correct
- Check key hasn't been revoked
- Generate new key if needed

## Related Documentation

- [Frontend Apps README](../README.md)
- [API Service README](../../functions/cmd/api/README.md)
- [Project Architecture](../../README.md)
