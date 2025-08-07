const url = 'http://localhost:8080/careers';

export async function fetchTrackerData(): Promise<any> {
    const res = await fetch(url, {
        method: 'GET',
        credentials: 'include',
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }

    return await res.json();
}

export async function updateTracker(frequency?:string, quantity?:number, penalty?:boolean): Promise<any> {
    const updateTracker = {
        cycleFrequency: frequency,
        goalQuantity: quantity,
        missedGoalPenalty: penalty,
    };
    const res = await fetch(url, {
        method: 'PATCH',
        credentials: 'include',
        body: JSON.stringify(updateTracker)
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }
    return await res.json();
}

export async function createTrackerItem(title: string, body: string): Promise<any> {
    const item = {
        title: title,
        body: body,
        status: "complete",
        isAttributed: false,
    };
    const res = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify(item)
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }

    return await res.json();
}

// TODO: implement status updates
export async function updateTrackerItem(id: string, title?: string, body?: string): Promise<any> {
    const item: Record<string, any> = { id };
    if (title) item.title = title;
    if (body) item.body = body;

    const res = await fetch(url + '?id=' + id, {
        method: 'PUT',
        credentials: 'include',
        body: JSON.stringify(item)
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }

    return await res.json();
}

export async function deleteTrackerItem(id: string): Promise<any> {
    const res = await fetch(url + '?id=' + id, {
        method: 'DELETE',
        credentials: 'include',
    });
    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }
    return await res.json();
}

export function formatDate(raw: string) {
    return new Date(raw);
}
