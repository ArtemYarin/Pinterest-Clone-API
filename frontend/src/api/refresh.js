import client from "./client";

// Exchanges the refresh_token cookie for a new access token: { token }.
export async function refresh() {
    const response = await client.post('/auth/refresh');
    return response.data;
}
