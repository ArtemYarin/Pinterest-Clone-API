import client from "./client";

// Returns pins liked by the user, newest like first. Same shape as getPins.
export async function likedPins(userId, {signal} = {}) {
    const response = await client.get(`/users/${userId}/liked-pins`, { signal });
    return response.data.data ?? [];
}
