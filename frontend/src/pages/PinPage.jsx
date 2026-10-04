import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router'
import axios from 'axios'
import { getPin } from '../api/getPin'
import usePins from '../hooks/usePins'
import PinGrid from '../features/PinGrid'
import LikeButton from '../components/LikeButton'

const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' })

function Message({ children }) {
  return (
    <div className='flex flex-1 items-center justify-center'>
      <p className='font-display text-5xl opacity-75'>{children}</p>
    </div>
  )
}

export default function PinPage() {
  const { pinId } = useParams()

  const [pin, setPin] = useState(null)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)
  const { pins, loading: pinsLoading, error: pinsError } = usePins('')

  useEffect(() => {
    const controller = new AbortController()

    window.scrollTo(0, 0)
    setPin(null)
    setError(null)
    setLoading(true)

    getPin(pinId, { signal: controller.signal })
      .then(setPin)
      .catch((err) => {
        if (axios.isCancel(err)) return
        setError(err.message)
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })

    return () => {
      controller.abort()
    }
  }, [pinId])

  const otherPins = pins.filter((p) => p.pin.id !== pinId)

  let content
  if (error) content = <Message>Error: {error}</Message>
  else if (loading || !pin) content = <Message>Loading pin…</Message>
  else {
    const { title, description, likes, created_at } = pin.pin
    content = (
      <article className='mx-auto flex w-full max-w-content flex-col overflow-hidden rounded-card bg-card md:flex-row'>
        <div className='flex items-center justify-center bg-bg-focus md:w-1/2'>
          {pin.download_url ? (
            <img
              className='max-h-[80dvh] w-full object-contain'
              src={pin.download_url}
              alt={title}
            />
          ) : (
            <p className='p-card opacity-60'>Image unavailable</p>
          )}
        </div>
        <div className='flex flex-col gap-4 p-card md:w-1/2'>
          <LikeButton pinId={pinId} initialCount={likes ?? 0} />
          <h1 className='font-display text-title font-bold wrap-break-word'>
            {title}
          </h1>
          {description && (
            <p className='whitespace-pre-line opacity-80'>{description}</p>
          )}
          {created_at && (
            <p className='mt-auto text-sm opacity-60'>
              Posted {dateFormat.format(new Date(created_at))}
            </p>
          )}
        </div>
      </article>
    )
  }

  return (
    <div className='flex min-h-dvh flex-col gap-8 px-6 pb-6'>
      <header className='sticky top-0 z-40 bg-bg py-4'>
        <Link
          to='/'
          className='inline-block rounded-button bg-card p-button font-semibold hover:bg-card-focus'
        >
          ← Back
        </Link>
      </header>

      {content}

      <section className='flex flex-col gap-4'>
        <h2 className='font-display text-big font-bold'>More pins</h2>
        <PinGrid pins={otherPins} loading={pinsLoading} error={pinsError} />
      </section>
    </div>
  )
}
