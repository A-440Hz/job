const url = 'http://localhost:8080/careers';

const timezoneOffset = new Date().getTimezoneOffset();

export async function fetchTrackerData(): Promise<any> {
    const headers = new Headers();
    headers.append('X-Timezone-Offset', timezoneOffset.toString());
    const res = await fetch(url, {
        method: 'GET',
        credentials: 'include',
        headers,
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }

    return await res.json();
}

export async function updateTracker(frequency?:string, deadline?:number, quantity?:number, penalty?:boolean): Promise<any> {
    const updateTracker: Record<string, any> = {};

    if (frequency) updateTracker.cycleFrequency = frequency;
    if (deadline) updateTracker.cycleDeadline = deadline;
    if (quantity) updateTracker.goalQuantity = quantity;
    if (penalty !== undefined) updateTracker.missedGoalPenalty = penalty;
    console.log("updateTracker:", updateTracker)
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

