# Frontend Applications

Cycle Scene consists of four SvelteKit/Vite applications for different user-facing features and admin functions.

## Applications Overview

### Directory
City-based landing page where users select their city to access the PWA.

**Location**: `/frontends/directory/`
**Purpose**: Entry point and city selection
**Stack**: SvelteKit, Svelte 5, TailwindCSS, LibSQL (analytics)
**URL**: https://cyclescene.cc

[Read Directory README](directory/README.md)

### PWA (Progressive Web App)
Mobile-first application for discovering, viewing, and saving bike rides.

**Location**: `/frontends/pwa/`
**Purpose**: Main user application
**Stack**: Svelte 5, Vite, MapLibre GL, Workbox, IndexedDB
**URL**: https://pdx.cyclescene.cc (and other cities)

[Read PWA README](pwa/README.md)

### Form
User interface for submitting new bike rides.

**Location**: `/frontends/form/`
**Purpose**: Ride submission
**Stack**: SvelteKit, Svelte 5, TailwindCSS, Zod validation
**URL**: https://form.cyclescene.cc

[Read Form README](form/README.md)

### Dashboard
Admin dashboard for viewing analytics and managing rides.

**Location**: `/frontends/dashboard/`
**Purpose**: Admin analytics and management
**Stack**: SvelteKit, Svelte 5, TailwindCSS
**URL**: https://dashboard.cyclescene.cc

[Read Dashboard README](dashboard/README.md)

## Shared Architecture

All applications use:
- Svelte 5 (reactive components)
- Vite (build tool)
- TailwindCSS (styling)
- TypeScript (type safety)

## Getting Started

### Prerequisites
- Node.js 18+
- npm or pnpm

### Setup All Apps
```bash
cd frontends
npm install
```

### Run Individual App
```bash
cd frontends/pwa
npm run dev
```

### Build for Production
```bash
npm run build
npm run preview
```

## Directory Structure

```
frontends/
├── directory/        # City directory
│   ├── src/
│   │   ├── routes/
│   │   ├── lib/
│   │   └── app.css
│   ├── package.json
│   └── README.md
├── pwa/             # Mobile app
│   ├── src/
│   │   ├── routes/
│   │   ├── lib/
│   │   └── app.css
│   ├── public/
│   │   └── manifest.json
│   ├── package.json
│   └── README.md
├── form/            # Ride submission
│   ├── src/
│   │   ├── routes/
│   │   ├── lib/
│   │   └── app.css
│   ├── package.json
│   └── README.md
├── dashboard/       # Admin
│   ├── src/
│   │   ├── routes/
│   │   ├── lib/
│   │   └── app.css
│   ├── package.json
│   └── README.md
├── package.json
└── README.md
```

## Development

### Project Structure Pattern

Each app follows this pattern:

**Routes** (`src/routes/`)
- Page components using SvelteKit's file-based routing
- API endpoints in `+server.ts` files
- Layout components in `+layout.svelte`

**Components** (`src/lib/components/`)
- Reusable UI components
- Custom component library (buttons, cards, forms)
- Feature-specific components

**Utilities** (`src/lib/`)
- Helper functions
- API client functions
- Stores and state management
- Type definitions

**Styling**
- TailwindCSS for styling
- Global styles in `app.css`
- Component-scoped styles with `<style>` blocks

### Common Tasks

#### Add New Page
1. Create file in `src/routes/newpage/+page.svelte`
2. Define page component with any needed logic
3. Add route to navigation if applicable

#### Add New Component
1. Create file in `src/lib/components/NewComponent.svelte`
2. Define component with props and events
3. Import and use in pages/layouts

#### Fetch Data from API
```typescript
// In +page.ts or component load function
async function fetchRides(city: string) {
  const response = await fetch(`/api/rides?city=${city}`);
  return response.json();
}
```

#### Use TailwindCSS
```svelte
<div class="flex gap-4 bg-slate-100 dark:bg-slate-900 p-4 rounded-lg">
  <p class="text-lg font-semibold">Hello World</p>
</div>
```

## Deployment

### Environment Variables

Each app uses environment variables for configuration. Create `.env` file in app directory:

```
VITE_API_URL=https://api.cyclescene.cc
VITE_APP_VERSION=1.0.0
```

Prefix with `VITE_` to expose to client-side code.

### Build

```bash
npm run build
```

Generates optimized production build in `build/` directory.

### Vercel Deployment

All apps are deployed to Vercel:

```bash
vercel deploy
```

Auto-deploys on git push to main branch.

### Docker (if needed)

```bash
docker build -t cyclescene-pwa:latest .
docker run -p 3000:3000 cyclescene-pwa:latest
```

## Styling

### TailwindCSS Configuration

Configuration is in `tailwind.config.ts`. Common customization:
- Color palette
- Font families
- Breakpoints
- Custom plugins

### Dark Mode

TailwindCSS dark mode is enabled. Use `dark:` prefix:

```svelte
<div class="bg-white dark:bg-slate-900">
  <!-- Light background on light mode, dark on dark mode -->
</div>
```

### Responsive Design

Mobile-first approach. Use breakpoints:
- `sm:` - 640px
- `md:` - 768px
- `lg:` - 1024px
- `xl:` - 1280px

```svelte
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
  <!-- 1 column on mobile, 2 on tablet, 3 on desktop -->
</div>
```

## State Management

### Simple State
Use Svelte stores for shared state:

```typescript
// src/lib/stores.ts
import { writable } from 'svelte/store';

export const rides = writable([]);
export const selectedCity = writable('pdx');
```

### Complex State
Use context API for feature-specific state:

```typescript
// In parent component
setContext('key', value);

// In child component
const value = getContext('key');
```

## API Integration

### Fetch Helper
Most apps have API helper in `src/lib/api.ts`:

```typescript
export async function getRides(city: string) {
  const response = await fetch(`/api/rides?city=${city}`);
  if (!response.ok) throw new Error('Failed to fetch rides');
  return response.json();
}
```

### Error Handling
Implement proper error handling:

```typescript
try {
  const rides = await getRides(city);
} catch (error) {
  console.error('Failed to load rides:', error);
  // Show error UI
}
```

## Testing

### Run Tests
```bash
npm test
```

### Write Tests
Create test files alongside components:

```
Button.svelte
Button.test.ts
```

## Performance

### Optimization Tips
- Use `{#await}` blocks for async data
- Lazy load images with `loading="lazy"`
- Code split pages with dynamic imports
- Monitor bundle size with `npm run analyze`

### Monitoring
Most apps include analytics (see Directory for example).

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers (iOS Safari 14+, Chrome Android)

## Contributing

When adding new features:
1. Create feature branch from main
2. Make changes and test locally
3. Push to GitHub
4. Vercel automatically deploys preview
5. After review, merge to main for production

## Troubleshooting

### Dev Server Won't Start
```bash
# Clear node_modules and reinstall
rm -rf node_modules
npm install
npm run dev
```

### Build Fails
```bash
# Check for TypeScript errors
npm run check

# Check for lint errors
npm run lint
```

### Styling Not Applied
- Verify TailwindCSS config includes your files
- Check class names are valid Tailwind classes
- Clear cache: `npm run build` then try again

## Related Documentation

- [Root README](../README.md)
- [Backend Services README](../functions/README.md)
- [SvelteKit Docs](https://kit.svelte.dev/)
- [Tailwind Docs](https://tailwindcss.com/docs)
