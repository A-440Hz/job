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