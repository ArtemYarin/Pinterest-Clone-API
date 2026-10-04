import { useCallback, useEffect, useRef, useState } from 'react'
import { addLike, getLikeCount, hasLiked, removeLike } from '../api/likes'
import { useAuth } from '../context/authContext'

export default function useLike(pinId, initialCount = 0) {
    const { isAuthenticated } = useAuth()
    const [liked, setLiked] = useState(false)
    const [count, setCount] = useState(initialCount)
    const [pending, setPending] = useState(false)
    const pendingRef = useRef(false)

    useEffect(() => {
        const controller = new AbortController()
        setCount(initialCount)

        getLikeCount(pinId, {signal: controller.signal})
            .then(setCount)
            .catch(() => {})

        return () => {
            controller.abort()
        }
    }, [pinId, initialCount])

    useEffect(() => {
        if (!isAuthenticated) {
            setLiked(false)
            return
        }
        const controller = new AbortController()

        hasLiked(pinId, {signal: controller.signal})
            .then(setLiked)
            .catch(() => {})

        return () => {
            controller.abort()
        }
    }, [pinId, isAuthenticated])

    const toggle = useCallback(async () => {
        if (pendingRef.current) return
        pendingRef.current = true
        setPending(true)

        const next = !liked
        // Optimistic update; roll back if the request fails.
        setLiked(next)
        setCount((c) => c + (next ? 1 : -1))
        try {
            await (next ? addLike(pinId) : removeLike(pinId))
        } catch {
            setLiked(!next)
            setCount((c) => c + (next ? -1 : 1))
        } finally {
            pendingRef.current = false
            setPending(false)
        }
    }, [liked, pinId])

    return {liked, count, toggle, pending}
}
