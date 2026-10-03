import client from "./client";

// Makes api call to login endpoint and returns a token.
export async function login(body) {
    const response = await client.post('/auth/login', body);
    return response.data;
}
