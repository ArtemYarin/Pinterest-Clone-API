import client from "./client";

// Returns an object: user_id, username, bio, avatar_url, created_at, updated_at
export async function getProfile(userId, {signal} = {}) {
    const response = await client.get(`/profile/${userId}`, { signal });
    return response.data;
}
