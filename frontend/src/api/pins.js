import client from "./client";

export async function getPins() {
    const response = await client.get("/pin");
    return response.data.data;
}