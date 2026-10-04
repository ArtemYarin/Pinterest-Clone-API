import client from "./client";

export async function getLikeCount(pinId, {signal} = {}) {
    const response = await client.get(`/interaction/${pinId}/like/count`, { signal });
    return response.data.count;
}

export async function hasLiked(pinId, {signal} = {}) {
    const response = await client.get(`/interaction/${pinId}/like`, { signal });
    return response.data.liked;
}

export async function addLike(pinId) {
    await client.put(`/interaction/${pinId}/like`);
}

export async function removeLike(pinId) {
    await client.delete(`/interaction/${pinId}/like`);
}
