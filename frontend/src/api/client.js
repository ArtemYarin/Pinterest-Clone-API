import axios from "axios";

// Requests go to the same origin; in dev Vite proxies them to the API (see vite.config.js).
const client = axios.create({
    baseURL: "/",
    withCredentials: true,
    headers: {
        "Content-Type": "application/json",
    },
});

export default client
