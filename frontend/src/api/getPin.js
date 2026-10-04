import client from "./client";

// Returns an object: download_url, pin
export async function getPin(id, {signal} = {}) {
    const response = await client.get(`/pin/${id}`, { signal });
    return response.data;
}