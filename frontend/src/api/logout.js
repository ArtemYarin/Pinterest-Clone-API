import client from "./client";

// Revokes the refresh token and clears its cookie.
export async function logout() {
    await client.post('/auth/logout');
}
