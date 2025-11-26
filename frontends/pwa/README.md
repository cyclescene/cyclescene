# PWA (Progressive Web App)

Mobile-first application for discovering, viewing, and saving bike rides in a specific city.

## Overview

The PWA is the main user application for browsing and interacting with bike rides. Users can:
- Browse rides on an interactive map
- View rides in list format
- Filter rides by date
- Save favorite rides locally
- Use the app offline with cached data
- Install as a native app on their device

**Technology**: Svelte 5, Vite, MapLibre GL, Workbox, IndexedDB
**URL**: https://pdx.cyclescene.cc (and other cities)
**Deployment**: Vercel

## Getting Started

### Prerequisites
- Node.js 18+
- Mapbox API key (for maps)

### Local Setup

```bash
cd frontends/pwa
npm install
npm run dev
```

Access at `http://localhost:5173`

## Features

### Map View
- Interactive map powered by MapLibre GL
- Pin drops for each ride
- Click ride markers for details
- Geolocation support to find rides near you

### List View
- Compact ride listing
- Sortable and filterable
- Shows distance if location available
- Quick access to ride details

### Saved Rides
- Save favorite rides locally (stored in IndexedDB)
- Access saved rides offline
- Quick reference for interested rides

### Date Filtering
- Pick date range to view rides
- See past and upcoming rides
- Filter by specific date

### Offline Support
- Service Worker caching for offline access
- Works without internet connection
- Syncs when connection is restored

### Install Prompt
- Appears on first visit
- Allows installation as native app
- Creates persistent local storage

## Project Structure

```
pwa/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte      # Main layout
│   │   ├── +page.svelte        # Map/list view
│   │   ├── [id]/
│   │   │   └── +page.svelte    # Ride details
│   │   ├── saved/
│   │   │   └── +page.svelte    # Saved rides
│   │   └── settings/
│   │       └── +page.svelte    # User settings
│   ├── lib/
│   │   ├── components/         # UI components
│   │   ├── api.ts              # API client
│   │   ├── store.ts            # State management
│   │   └── db.ts               # IndexedDB operations
│   └── app.css
├── public/
│   ├── manifest.json           # PWA manifest
│   ├── service-worker.js       # Offline support
│   └── icons/                  # App icons
├── package.json
└── README.md
```

## Data & Storage

### API Endpoints
- `GET /api/rides?city={code}` - Get all rides for a city
- `GET /api/rides/{id}` - Get detailed ride information

### Local Storage
IndexedDB tables:
- `rides` - Cached ride data
- `saved_rides` - User's saved rides
- `settings` - User preferences

Service Worker caching:
- Static assets (HTML, CSS, JS)
- API responses
- Images for offline viewing

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

### Environment Variables
- `VITE_API_URL` - API base URL
- `VITE_MAPBOX_TOKEN` - Mapbox API key

## Planned Features

### Install Prompt Enhancement
Add prominent "Install App" button for first-time users with better UX.

### Settings Expansion
- Units preference (miles/km)
- Show/hide past rides
- Notification preferences
- Dark mode toggle

### Calendar Export
Export rides to calendar format (ICS).

## Performance

### Optimizations
- Code splitting for routes
- Lazy loading images
- Service Worker caching
- Minimal bundle size

### Monitoring
Monitor in Google Analytics:
- Page load times
- User interactions
- Offline usage patterns

## Troubleshooting

### Map Not Showing
- Verify `VITE_MAPBOX_TOKEN` environment variable is set
- Check browser console for errors
- Ensure Mapbox account has sufficient credits

### Offline Not Working
- Verify service worker is registered in DevTools
- Check Cache Storage is enabled in browser
- Clear cache and reload

### Slow Performance
- Profile with DevTools Performance tab
- Check for unnecessary re-renders in Svelte components
- Optimize large images before upload

### Save Feature Not Working
- Verify IndexedDB is enabled in browser
- Check browser storage quota
- Clear cache and try again

## Browser Support
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers (iOS Safari 14+, Chrome Android)

## Related Documentation

- [Frontend Apps README](../README.md)
- [API Service README](../../functions/cmd/api/README.md)
- [Directory README](../directory/README.md)
- [Project Architecture](../../README.md)
