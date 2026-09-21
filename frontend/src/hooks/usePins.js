import { useState, useEffect } from 'react'
import { getPins } from '../api/pins'

export default function usePins(query = '') {
    const [pins, setPins] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        const controller = new AbortController()
        setError(null)
        setLoading(true)

        getPins(query, {signal: controller.signal})
            .then(setPins)
            .catch((err) => {
                if (err.name === 'AbortError' || err.name === 'CanceledError' || err.message === 'canceled') {
                    return
                }
                setError(err.message)
            })
            .finally(() => {
                if (!controller.signal.aborted) setLoading(false)
            })

        return () => {
            controller.abort()
        }
    }, [query])

    return {pins, loading, error}
}