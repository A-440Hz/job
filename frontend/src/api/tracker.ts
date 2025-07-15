
export async function fetchTrackerData(): Promise<any> {
    const res = await fetch('http://localhost:8080/careers', {
        method: 'GET',
        credentials: 'include',
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }

    return await res.json();
}