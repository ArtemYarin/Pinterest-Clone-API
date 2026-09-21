import client from "./client";

export async function getPins(query, {signal} = {}) {
    const response = await client.get(`/pin?search=${query}`, {signal: signal});
    return response.data.data;
}