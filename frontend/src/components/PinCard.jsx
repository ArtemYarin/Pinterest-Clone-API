export default function PinCard({ pin }) {
  return (
    <div className='mb-1 break-inside-avoid'>
      <img
        src={pin.download_url}
        alt={pin.pin.title}
        className='w-full rounded-xl block'
      />
    </div>
  )
}
