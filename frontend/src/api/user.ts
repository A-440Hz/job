import { endpoint } from "./endpoint"

const registerEndpoint = endpoint+'/careers/register'
const loginEndpoint = endpoint+'/careers/login'
const logoutEndpoint = endpoint+'/careers/logout'

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

