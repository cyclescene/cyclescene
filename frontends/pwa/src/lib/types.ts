export interface RideData {
  id: string
  title: string
  lat: number
  lng: number
  address: string
  audience: string
  cancelled: number
  date: string
  starttime: string
  safetyplan: number
  details: string
  venue: string
  organizer: string
  loopride: number
  shareable: string
  ridesource: string
  route_id?: string | null
  endtime: string
  email: string
  eventduration: number
  image: string
  locdetails: string
  locend: string
  newsflash: string
  timedetails: string
  weburl: string
  webname: string
  group_code?: string
  group_marker?: string
}


export interface ValidatedRide {
  id: string;
  name: string;
  lat: number;
  lng: number;
  marker_key?: string;
}


export interface ShiftEvent {
  id: string;
  title: string;
  venue: string;
  address: string;
  organizer: string;
  details: string;
  time: string;
  hideemail: string;
  length?: null;
  timedetails?: null;
  locdetails?: null;
  eventduration: string;
  weburl?: null;
  webname: string;
  image: string;
  audience: string;
  tinytitle: string;
  printdescr: string;
  datestype: string;
  area: string;
  featured: boolean;
  printemail: boolean;
  printphone: boolean;
  printweburl: boolean;
  printcontact: boolean;
  email?: null;
  phone?: null;
  contact?: null;
  date: string;
  caldaily_id: string;
  shareable: string;
  cancelled: boolean;
  newsflash?: null;
  status: string;
  endtime: string;
}

export type ShiftEventResponse = {
  events: ShiftEvent[]
}



