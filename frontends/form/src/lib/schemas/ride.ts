import { z } from 'zod'

// Occurrence schema for individual ride dates
export const rideOccurrenceSchema = z.object({
  start_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format'),
  start_time: z.string().regex(/^\d{2}:\d{2}:\d{2}$/, 'Invalid time format'),
  event_duration_minutes: z.number().int().min(0).optional(),
  event_time_details: z.string().optional(),
  newsflash: z.string().max(500).optional()
});

// Main ride submission schema
export const rideSubmissionSchema = z.object({
  // Core content
  title: z.string().min(3, 'Title must be at least 3 characters').max(200),
  tinytitle: z.string().max(50).optional(),
  description: z.string().min(10, 'Description must be at least 10 characters'),
  image_url: z.union([z.httpUrl(), z.literal('')]).optional(),
  image_uuid: z.string().optional(),
  audience: z.enum(['G', 'F', 'A', 'E'], {
    error: () => ({ message: 'Please select an audience type' })
  }).optional(),
  ride_length: z.string().max(50).optional(),
  area: z.string().max(100).optional(),
  date_type: z.enum(['S', 'R', 'O'], {
    error: () => ({ message: 'Please select a date type' })
  }),

  // Location
  venue_name: z.string().min(2, 'Venue name is required').max(200),
  address: z.string().min(5, 'Address is required'),
  location_details: z.string().optional(),
  ending_location: z.string().optional(),
  is_loop_ride: z.boolean().default(false),

  // Contact info
  organizer_name: z.string().min(2, 'Organizer name is required').max(100),
  organizer_email: z.email('Must be a valid email'),
  organizer_phone: z.string()
    .refine(val => !val || /^[\d\-\+\(\)\s]+$/.test(val), 'Invalid phone number format')
    .optional(),
  web_url: z.union([z.httpUrl(), z.literal('')]).optional(),
  web_name: z.string().max(100).optional(),
  newsflash: z.string().max(500).optional(),
  hide_email: z.boolean().default(false),
  hide_phone: z.boolean().default(false),
  hide_contact_name: z.boolean().default(false),

  // Group association
  group_code: z.string().length(4, 'Group code must be 4 characters').toUpperCase().optional().or(z.literal('')),

  // City (will be set from referrer/token)
  city: z.string().min(2, 'City is required'),

  // Occurrences
  occurrences: z.array(rideOccurrenceSchema).min(1, 'At least one date is required')
});

export type RideSubmission = z.infer<typeof rideSubmissionSchema>;
export type RideOccurrence = z.infer<typeof rideOccurrenceSchema>;

// Group registration schema
export const groupRegistrationSchema = z.object({
  code: z.string()
    .length(4, 'Group code must be exactly 4 characters')
    .regex(/^[A-Z0-9]+$/, 'Code must contain only letters and numbers')
    .toUpperCase(),
  name: z.string().min(3, 'Group name must be at least 3 characters').max(100),
  description: z.string().max(500).optional(),
  city: z.string().min(2, 'City is required'),
  email: z.string().email('Must be a valid email address'),
  image_uuid: z.string().optional(),
  web_url: z.union([z.httpUrl(), z.literal('')]).optional()
});

export type GroupRegistration = z.infer<typeof groupRegistrationSchema>;

// Audience options for the form
export const audienceOptions = [
  { value: 'G', label: 'General - All ages and abilities' },
  { value: 'F', label: 'Family-Friendly - Great for kids' },
  { value: 'A', label: 'Adults Only - 21+' },
  { value: 'E', label: 'Experienced - Advanced riders' }
] as const;

// Date type options
export const dateTypeOptions = [
  { value: 'S', label: 'Single Date - One-time event' },
  { value: 'R', label: 'Recurring - Repeats regularly' },
  { value: 'O', label: 'One-Off - Special event' }
] as const;

