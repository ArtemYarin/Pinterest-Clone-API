import axios from "axios";
import client from "./client";

export async function getPins(query, {signal} = {}) {
    const response = await client.get('/pin', {params: {search: query}, signal});
    return response.data.data ?? [];
}

// Creates the pin metadata. Returns an object: pin, upload_url (presigned PUT, 15 min).
export async function createPin(body) {
    const response = await client.post('/pin', body);
    return response.data;
}

// Uploads the image straight to object storage. Uses plain axios, not `client`:
// an Authorization header or credentials would break the presigned request.
// Content-Type is stored with the object and checked on confirm.
export async function uploadPinImage(uploadUrl, file, {onProgress} = {}) {
    await axios.put(uploadUrl, file, {
        headers: {"Content-Type": file.type},
        onUploadProgress: (e) => {
            if (onProgress && e.total) onProgress(Math.round((e.loaded / e.total) * 100));
        },
    });
}

// Validates the uploaded image and marks the pin as confirmed.
export async function confirmPinUpload(id) {
    const response = await client.post(`/pin/${id}/confirm`);
    return response.data;
}

export async function deletePin(id) {
    await client.delete(`/pin/${id}`);
}
