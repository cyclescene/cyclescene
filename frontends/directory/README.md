# Directory

City-based directory for selecting which city's PWA to access. This is the landing page for Cycle Scene.

## Overview

The Directory serves as the main entry point where users:
1. Land on cyclescene.cc
2. See available cities
3. Select their city
4. Get routed to the city-specific PWA

**Technology**: SvelteKit, Svelte 5, TailwindCSS, LibSQL (analytics)
**URL**: https://cyclescene.cc
**Deployment**: Vercel

## Getting Started

### Prerequisites
- Node.js 18+
- TursoDB account (for analytics database)

### Local Setup

1. Navigate to directory:
```bash
cd frontends/directory
```

2. Install dependencies:
```bash
npm install
```

3. Set environment variables:
```bash
# Create .env file
export ANALYTICS_DB_URL="libsql://your-database-url"
export ANALYTICS_DB_TOKEN="your-auth-token"
```

4. Run dev server:
```bash
npm run dev
```

Access at `http://localhost:5173`

## Features

### City Selection
- Browse available cities with emoji banners
- Click to navigate to city PWA
- Responsive grid layout for all screen sizes

### Analytics Tracking
- Tracks page visits (non-invasive)
- Records marketing source from URL parameters
- Tracks which city users click on
- Uses separate analytics database (no user data stored)

### Privacy
- Transparent privacy policy at `/privacy`
- Footer notice explaining analytics
- No IP tracking or cookies
- No login required

### Styling
- Modern, minimal design
- Dark mode support
- Mobile-responsive
- TailwindCSS 4

## Project Structure

```
directory/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte      # Main layout with footer
│   │   ├── +page.svelte        # Landing page with city selector
│   │   ├── privacy/
│   │   │   └── +page.svelte    # Privacy policy page
│   │   └── api/
│   │       └── analytics/
│   │           └── +server.ts  # Analytics API
│   ├── lib/
│   │   ├── data/
│   │   │   └── cities.json     # City configuration
│   │   ├── components/
│   │   │   └── EmojiBanner.svelte
│   │   └── server/
│   │       └── analytics-db.ts
│   └── app.css
├── package.json
└── README.md
```

## City Configuration

Cities are configured in `src/lib/data/cities.json`:

```json
[
  {
    "name": "Portland",
    "code": "pdx",
    "url": "https://pdx.cyclescene.cc"
  },
  {
    "name": "Salt Lake City",
    "code": "slc",
    "url": "https://slc.cyclescene.cc"
  }
]
```

To add a new city:
1. Add entry to cities.json
2. Deploy new city PWA with matching code
3. Redeploy directory

## Analytics

### Data Collected

On page visit:
- Page visit timestamp
- Marketing source (from `?source=` URL parameter)
- Browser information (user-agent, accept-language)

On city click:
- Which city was clicked
- Click timestamp

### Database

Analytics stored in separate TursoDB instance:

Table: `directory_analytics`
```sql
- id: Unique record ID
- source: Marketing campaign ID (e.g., "qr-sticker-pdx")
- clicked_cta: Whether user clicked city (0 or 1)
- pwa_clicked: City code clicked (e.g., "pdx")
- request_headers: Non-invasive headers (JSON)
- visited_at: Visit timestamp
- clicked_at: Click timestamp (if applicable)
```

### Privacy

Analytics are transparent:
- Privacy policy discloses what's collected
- Footer notice on every page
- Headers filtered to exclude IP addresses
- No cookies or user tracking
- Data stored only on our servers

### Marketing Source Tracking

Use `?source=` parameter to track campaigns:

```
https://cyclescene.cc?source=qr-sticker-pdx
https://cyclescene.cc?source=twitter-organic
https://cyclescene.cc?source=reddit-post
```

Then analyze which sources drive most traffic.

## Development

### Add New City

1. Update `src/lib/data/cities.json`
2. Deploy/ensure city PWA is live
3. Commit and push changes
4. Vercel auto-deploys

### Modify Landing Page

Edit `src/routes/+page.svelte`:
- Change heading/description
- Modify city grid layout
- Update button styling

### Update Privacy Policy

Edit `src/routes/privacy/+page.svelte`:
- Keep it clear and concise
- Update last modified date
- Ensure accuracy for your data practices

## Deployment

### Local Build
```bash
npm run build
npm run preview
```

### Vercel Deployment
```bash
vercel deploy
```

Auto-deploys on git push to main branch.

### Environment Variables (Vercel)

Set in Vercel project settings:
- `ANALYTICS_DB_URL`
- `ANALYTICS_DB_TOKEN`

## Features to Add

### Install Prompt
Show "Install App" button on first visit to encourage PWA installation.

Implementation: IndexedDB to track first visit, show modal with install button.

### Push Notifications
Notify users when major rides are added to their city.

### Newsletter Signup
Allow users to subscribe to city-specific ride updates.

## Monitoring

### Analytics Dashboard
View analytics in TursoDB console:

```sql
SELECT source, COUNT(*) as visits, SUM(clicked_cta) as clicks
FROM directory_analytics
GROUP BY source;

SELECT pwa_clicked, COUNT(*) as clicks
FROM directory_analytics
WHERE clicked_cta = 1
GROUP BY pwa_clicked;
```

### Traffic Sources
See which campaigns drive the most traffic:
```sql
SELECT source, COUNT(*) as visitors
FROM directory_analytics
WHERE visited_at > DATE('now', '-7 days')
GROUP BY source
ORDER BY visitors DESC;
```

### City Popularity
See which cities are most popular:
```sql
SELECT pwa_clicked, COUNT(*) as clicks
FROM directory_analytics
WHERE pwa_clicked IS NOT NULL
GROUP BY pwa_clicked
ORDER BY clicks DESC;
```

## Styling

### Colors
Uses TailwindCSS default colors with dark mode support:
- Light: White, slate-100
- Dark: slate-900

### Typography
- Heading: text-6xl, font-bold
- Subheading: text-xl
- Body: text-base
- Footer: text-xs

### Responsive
- Mobile: Single column
- Tablet: 2 columns (md breakpoint)
- Desktop: 2 columns (maintains readability)

## Performance

### Optimizations
- No external dependencies (except Svelte)
- Analytics lazy-loads on client mount
- Static city data (no API calls needed)
- Minimal JavaScript (landing page priority)

### Page Speed
- First Contentful Paint: <1s
- Largest Contentful Paint: <2s
- Cumulative Layout Shift: <0.1

## Browser Support
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers (iOS Safari 14+, Chrome Android)

## Troubleshooting

### Analytics Not Tracking
- Verify `ANALYTICS_DB_URL` and `ANALYTICS_DB_TOKEN` are set
- Check browser console for errors
- Ensure TursoDB instance is running
- Try restarting dev server

### Cities Not Loading
- Check `src/lib/data/cities.json` is valid JSON
- Verify city URLs are correct and reachable
- Check browser network tab for 404s

### Styling Not Applied
- Verify TailwindCSS is configured
- Check `app.css` is imported in `+layout.svelte`
- Clear build cache: `npm run build`

## Related Documentation

- [Frontend Apps README](../README.md)
- [Backend Services README](../../functions/README.md)
- [PWA README](../pwa/README.md)
- [Project Architecture](../../README.md)
