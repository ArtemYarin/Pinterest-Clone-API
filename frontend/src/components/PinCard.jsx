import { Link } from 'react-router'

export default function PinCard({ pin }) {
  const pinId = pin.pin.id
  return (
    <Link to={`/${pinId}`}>
      <div className='mb-6 break-inside-avoid hover:brightness-60'>
        <img
          src={pin.download_url}
          alt={pin.pin.title}
          loading='lazy'
          className='w-full rounded-xl block'
        />
      </div>
    </Link>
  )
}
