export const CITIES = [
  { code: 'all', name: 'All Cities', state: '' },
  { code: 'pdx', name: 'Portland', state: 'OR' },
  { code: 'slc', name: 'Salt Lake City', state: 'UT' }
] as const;

export type CityCode = typeof CITIES[number]['code'];
