import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router'
import axios from 'axios'
import PinGrid from '../features/PinGrid'
import Avatar from '../components/Avatar'
import { useAuth } from '../context/authContext'
import { getProfile } from '../api/getProfile'
import { getUserPins } from '../api/pins'
import { likedPins } from '../api/likedPins'

const fetchers = { created: getUserPins, liked: likedPins }

// Public profile of any user: avatar, username and their created or liked pins.
export default function ProfilePage() {
  const { userId } = useParams()
  const { userId: myId } = useAuth()
  const isOwn = myId === userId

  const [profile, setProfile] = useState(null)

  // 'created' | 'liked'
  const [option, setOption] = useState('created')
  const [pins, setPins] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const controller = new AbortController()
    setProfile(null)

    getProfile(userId, { signal: controller.signal })
      .then(setProfile)
      .catch(() => {})

    return () => controller.abort()
  }, [userId])

  useEffect(() => {
    const controller = new AbortController()
    setPins([])
    setError(null)
    setLoading(true)

    fetchers[option](userId, { signal: controller.signal })
      .then(setPins)
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
  }, [userId, option])

  const tabClass = (value) =>
    `rounded-button p-button font-semibold ${
      option === value ? 'bg-text text-bg' : 'hover:bg-card-focus'
    }`

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

      <section className='flex flex-col items-center gap-2'>
        <div className='[&>img]:size-24'>
          <Avatar url={profile?.avatar_url} />
        </div>
        <h1 className='font-display text-title font-bold wrap-break-word'>
          {profile?.username || 'User'}
        </h1>
        {profile?.bio && <p className='opacity-80'>{profile.bio}</p>}
      </section>

      <section className='flex flex-1 flex-col gap-4'>
        <div className='flex justify-center gap-2'>
          <button
            className={tabClass('created')}
            type='button'
            onClick={() => setOption('created')}
          >
            {isOwn ? 'Your pins' : 'Created'}
          </button>
          <button
            className={tabClass('liked')}
            type='button'
            onClick={() => setOption('liked')}
          >
            Liked
          </button>
        </div>

        <PinGrid pins={pins} loading={loading} error={error} />
      </section>
    </div>
  )
}
