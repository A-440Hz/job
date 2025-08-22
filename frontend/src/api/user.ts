import { endpoint } from "./endpoint"

const registerEndpoint = endpoint+'/careers/register'

export async function registerUser(username: string, password: string, email?: string) {
    const req: Record<string, any> = { username, password };
    if (email) {
        req.email = email
    }
    const res = await fetch(registerEndpoint, {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify(req)
    });

    if (!res.ok) {
        throw new Error(`Failed to Fetch tracker: ${res.statusText}`);
    }

    return await res.json();
}