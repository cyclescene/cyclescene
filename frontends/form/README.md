# Form

User interface for submitting new bike rides.

## Overview

The Form application allows community members to submit bike rides to be displayed across Cycle Scene. Features include:
- Detailed ride submission form
- Date and time selection
- Location picker
- Image uploads
- Group selection
- Shift2Bikes modal redirect (for Portland)

**Technology**: SvelteKit, Svelte 5, TailwindCSS, Zod validation
**URL**: https://form.cyclescene.cc
**Deployment**: Vercel

## Getting Started

```bash
cd frontends/form
npm install
npm run dev
```

Access at `http://localhost:5173`

## Form Fields

- Title: Ride name
- Description: Detailed description
- Date & Time: Start and end times
- Location: Address or coordinates
- Group: Ride organizer
- Images: Optional ride photos
- Category: Type of ride

## Validation

Form uses Zod for schema validation:
- Required fields validation
- Date/time format validation
- Location geocoding
- File type validation for images

## Shift2Bikes Integration

For Portland rides, shows modal encouraging users to submit to Shift2Bikes instead (with "don't show again" option).

## API Integration

Submits to: `POST /api/rides`

Returns submission token for tracking.

## Styling

- TailwindCSS for styling
- Dark mode support
- Responsive form layout
- Accessibility features (labels, ARIA)

## Troubleshooting

### Form Not Submitting
- Check browser console for validation errors
- Verify API endpoint is accessible
- Ensure required fields are filled

### Images Not Uploading
- Check file size (max 10MB)
- Verify image format (JPEG, PNG, WebP)
- Check browser file upload permissions

## Related Documentation

- [Frontend Apps README](../README.md)
- [API Service README](../../functions/cmd/api/README.md)
- [Project Architecture](../../README.md)
