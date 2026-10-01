export default function PinCard({ pin }) {
  return (
    <div className='mb-6 break-inside-avoid hover:brightness-60'>
      <img
        src={pin.download_url}
        alt={pin.pin.title}
        className='w-full rounded-xl block'
      />
    </div>
  )
}
