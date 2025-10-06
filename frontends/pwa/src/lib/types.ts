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
}


export interface ValidatedRide {
  id: string;
  name: string;
  lat: number;
  lng: number;
}
