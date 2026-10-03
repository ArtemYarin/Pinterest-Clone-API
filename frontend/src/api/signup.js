import client from "./client";

// Makes api call to signup endpoint and returns a token with user data.
export async function signup(body) {
    const response = await client.post('/auth/signup', body);
    return response.data;
}
