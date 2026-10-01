import axios from "axios";

const host = import.meta.env.VITE_API_HOST;
const port = import.meta.env.VITE_API_PORT;

const client = axios.create({
    baseURL: `http://${host}:${port}/`,
    headers: {
        "Content-Type": "application/json",
    },
});

export default client