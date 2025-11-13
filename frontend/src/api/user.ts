import { endpoint } from "./endpoint"

const registerEndpoint = endpoint+'/careers/register'
const loginEndpoint = endpoint+'/careers/login'
const logoutEndpoint = endpoint+'/careers/logout'
const openLootboxEndpoint = endpoint+'/careers/profile/open'
const collectablesEndpoint = endpoint+'/careers/profile'

export async function fetchCollectablesData(): Promise<any> {
    const headers = new Headers();
    const res = await fetch(collectablesEndpoint, {
        method: 'GET',
        credentials: 'include',
        headers: headers,
        cache: 'no-cache',
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch collectables: ${res.statusText}`);
    }

    return await res.json();
}

export async function registerUser(username: string, password: string, email?: string) {
    const req: Record<string, any> = { username, password };
    if (email) {
        req.email = email
    }
    const res = await fetch(registerEndpoint, {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify(req),
    });

    if (!res.ok) {
        throw new Error(`Failed to register: ${res.statusText}`);
    }

    return await res.json();
}

export async function loginUser(username: string, password: string) {
    const req: Record<string, any> = { username, password };
    const res = await fetch(loginEndpoint, {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify(req),
    });
    if (!res.ok) {
        throw new Error(`Failed to login: ${res.statusText}`);
    }
    return await res.json();
}

export async function logoutUser() {
    const res = await fetch(logoutEndpoint, {
        method: 'POST',
        credentials: 'include',
    });
    if (!res.ok) {
        throw new Error(`Failed to logout: ${res.statusText}`);
    }
    return await res.json();
}

export async function openOneLootbox(): Promise<any> {
    const req = {award_ten: false};
    const res = await fetch(openLootboxEndpoint, {
        method: 'POST',
        credentials: 'include',
        cache: 'no-cache',
        body: JSON.stringify(req),
    });
    if (!res.ok) {
        throw new Error(`Failed to open lootbox: ${res.statusText}`);
    }
    const data = await res.json();
    console.log("opened one lootbox:", data);
    return data;
}

export async function openTenLootboxes(): Promise<any> {
    const req = {award_ten: true};
    const res = await fetch(openLootboxEndpoint, {
        method: 'POST',
        credentials: 'include',
        cache: 'no-cache',
        body: JSON.stringify(req),
    });
    if (!res.ok) {
        throw new Error(`Failed to open lootboxes: ${res.statusText}`);
    }
    const data = await res.json();
    console.log("opened ten lootboxes:", data);
    return data;
}
