import { createEvent } from 'ics';
import type { RideData } from './types';
// Assuming createEvent is imported from your local node_modules via Vite/TS setup

function getICSDateArray(date: string, time: string): [number, number, number, number, number] {
  const d = new Date(`${date} ${time}`);

  // Month is 1-indexed for the ics package, so we use getMonth() + 1
  return [
    d.getFullYear(),
    d.getMonth() + 1,
    d.getDate(),
    d.getHours(),
    d.getMinutes()
  ];
}

export async function downloadCalendarFile(ride: RideData) {

  // --- 1. Data Preparation ---
  const rideDate = ride.date;
  const startTime = ride.starttime || '00:00:00';
  let endTime = ride.endtime || '';

  // Calculate End Time (2 hours estimated duration if missing)
  if (!endTime) {
    const startDateTime = new Date(`${rideDate} ${startTime}`);
    const estimatedEndDateTime = new Date(startDateTime.getTime() + (2 * 60 * 60 * 1000));
    // Use 24-hour time for the internal calculation
    endTime = `${estimatedEndDateTime.getHours()}:${estimatedEndDateTime.getMinutes()}:${estimatedEndDateTime.getSeconds()}`;
  }

  // --- 2. Create the ICS Event Object ---
  const event = {
    // Use the array format for start/end time
    start: getICSDateArray(rideDate, startTime),
    end: getICSDateArray(rideDate, endTime),
    title: ride.title,
    description: `Join the ride! Details: ${ride.details.replace(/[\r\n]/g, '\n')}\nLink: ${ride.shareable}`,
    location: `${ride.venue || ride.address}`,
    url: ride.shareable,
    productId: 'CycleScene',
    uid: `${ride.id}@cyclescene.com`
  };

  const filename = `${ride.title.replace(/[\s\W]+/g, '_')}_${ride.date}.ics`;

  // --- 3. Run Asynchronous Event Creation ---
  const file = await new Promise<File>((resolve, reject) => {
    createEvent(event, (error, value) => {
      if (error) {
        console.error("ICS Creation Error:", error);
        // Return a simple, empty file on error to avoid crashing the flow
        reject(new Error("Failed to create ICS file."));
        return;
      }
      // Resolve the Promise with the File object
      resolve(new File([value], filename, { type: 'text/calendar' }));
    });
  }).catch(e => {
    console.error(e);
    // Fallback or re-throw after logging
    return null;
  });

  if (!file) return;

  // --- 4. Trigger Download (The HTML5/Browser Workaround) ---
  const url = URL.createObjectURL(file);
  const anchor = document.createElement('a');

  anchor.href = url;
  anchor.download = filename;

  // CRITICAL: Attach, click, and remove for cross-browser download reliability
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);

  // 5. Clean up
  URL.revokeObjectURL(url);
}
