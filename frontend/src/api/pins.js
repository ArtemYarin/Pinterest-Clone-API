import client from "./client";

export async function getPins(query, {signal} = {}) {
    const response = await client.get('/pin', {params: {search: query}, signal});
    return response.data.data ?? [];
}
