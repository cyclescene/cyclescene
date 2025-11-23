# Directory - CycleScene Landing Page

Simple static landing page that serves as a directory for all deployed CycleScene PWA instances.

## Features

- Reads city list from JSON configuration file
- Static site generation (no server required)
- Simple, scalable design
- Easy to add new cities

## Adding New Cities

To add a new city, simply add an object to `src/lib/data/cities.json`:

```json
{
  "name": "New City Name",
  "code": "city-code",
  "url": "https://city-code.cyclescene.cc"
}
```

Then rebuild and deploy:

```sh
pnpm build
```

## Development

```sh
pnpm install
pnpm run dev
```

Open [http://localhost:5173](http://localhost:5173) in your browser.

## Building

```sh
pnpm build
```

The static site will be generated in the `build/` directory.

## Deployment

Since this is a static site, it can be deployed to any static hosting service:
- Vercel
- Netlify
- GitHub Pages
- AWS S3 + CloudFront
- Any web server

Simply copy the `build/` directory contents to your hosting.
