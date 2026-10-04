import { useState } from 'react'
import { useAuth } from '../context/authContext'
import useLike from '../hooks/useLike'
import LoginModal from './auth/LoginModal'

function HeartIcon({ filled }) {
  return (
    <svg
      viewBox='0 0 24 24'
      className={`size-6 ${filled ? 'fill-accent stroke-accent' : 'fill-none stroke-current'}`}
      strokeWidth='2'
      strokeLinecap='round'
      strokeLinejoin='round'
      aria-hidden='true'
    >
      <path d='M12 21s-7.5-4.6-9.6-9.2C.9 8.4 3 4.5 6.7 4.5c2.1 0 3.5 1.1 5.3 3 1.8-1.9 3.2-3 5.3-3 3.7 0 5.8 3.9 4.3 7.3C19.5 16.4 12 21 12 21z' />
    </svg>
  )
}

export default function LikeButton({ pinId, initialCount }) {
  const { isAuthenticated, isInitializing } = useAuth()
  const { liked, count, toggle, pending } = useLike(pinId, initialCount)
  const [loginOpen, setLoginOpen] = useState(false)

  function handleClick() {
    // Wait for the session restore so a logged-in user isn't shown the login modal.
    if (isInitializing) return
    if (isAuthenticated) toggle()
    else setLoginOpen(true)
  }

  return (
    <>
      <button
        type='button'
        onClick={handleClick}
        disabled={pending}
        aria-pressed={liked}
        aria-label={liked ? 'Unlike pin' : 'Like pin'}
        className='flex w-fit items-center gap-2 rounded-button bg-bg p-button font-semibold transition-colors hover:bg-bg-focus disabled:opacity-60'
      >
        <HeartIcon filled={liked} />
        <span>{count}</span>
      </button>
      {loginOpen && <LoginModal onClose={() => setLoginOpen(false)} />}
    </>
  )
}
