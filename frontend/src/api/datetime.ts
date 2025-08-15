export function formatDate(raw: string) {
    return new Date(raw);
}

// dateToInputString converts a Date object to a string format for use with the <input> datetime-local element
// If I wanted to be more pedantic and safe style-wise, I can construct my own string with d.getFullYear()/..Month()/..Day()...
// it is good enough for me to assume toISOString output format will never change.
export function dateToInputString(d: Date) {
    const st = d.toISOString()
    return st.substring(0, st.indexOf('T')+6)
}

export function inputStringToDate(s: string) {
    return new Date(s+':00.000Z')
}

export function adjustTimezoneOffset(date:Date, seconds:number, add:boolean) {
  let offset_ms = seconds * 1000
  if (!add) {offset_ms = offset_ms * -1}
  return new Date(date.valueOf()+(offset_ms))
}

// convertToBackendTime converts the milliseconds UTC representation to a seconds input which feeds easily into a go time.Unix format 
export function convertToBackendTime(d:Date) {
    return Math.floor(d.valueOf()/1000)
}
