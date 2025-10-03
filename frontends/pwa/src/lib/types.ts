export interface RideData {
  address: string
  audience: string
  cancelled: number
  date: string
  details: Details
  email: Email
  endtime: Endtime
  eventduration: Eventduration
  id: string
  image: Image
  lat: Lat
  locdetails: Locdetails
  locend: Locend
  lon: Lon
  loopride: number
  newsflash: Newsflash
  organizer: Organizer
  safetyplan: number
  shareable: string
  source_data: string
  starttime: string
  timedetails: Timedetails
  title: string
  venue: Venue
  webname: Webname
  weburl: Weburl
}

export interface Details {
  String: string
  Valid: boolean
}

export interface Email {
  String: string
  Valid: boolean
}

export interface Endtime {
  String: string
  Valid: boolean
}

export interface Eventduration {
  Int32: number
  Valid: boolean
}

export interface Image {
  String: string
  Valid: boolean
}

export interface Lat {
  Float64: number
  Valid: boolean
}

export interface Locdetails {
  String: string
  Valid: boolean
}

export interface Locend {
  String: string
  Valid: boolean
}

export interface Lon {
  Float64: number
  Valid: boolean
}

export interface Newsflash {
  String: string
  Valid: boolean
}

export interface Organizer {
  String: string
  Valid: boolean
}

export interface Timedetails {
  String: string
  Valid: boolean
}

export interface Venue {
  String: string
  Valid: boolean
}

export interface Webname {
  String: string
  Valid: boolean
}

export interface Weburl {
  String: string
  Valid: boolean
}


export interface ValidatedRide {
  id: string;
  name: string;
  lat: number;
  lng: number;
}
