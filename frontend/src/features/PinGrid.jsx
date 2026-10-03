import PinCard from '../components/PinCard'

function Message({ children }) {
  return (
    <div className='flex flex-1 items-center justify-center'>
      <p className='font-display text-5xl opacity-75'>{children}</p>
    </div>
  )
}

export default function PinGrid({ pins, loading, error }) {
  if (error) return <Message>Error: {error}</Message>
  // Only replace the grid on the first load; later searches keep the old pins visible.
  if (loading && !pins.length) return <Message>Loading pins…</Message>
  if (!pins.length) return <Message>No pins</Message>

  return (
    <div
      className={`flex-1 columns-2 gap-6 transition-opacity md:columns-3 lg:columns-4 ${loading ? 'opacity-60' : ''}`}
      aria-busy={loading}
    >
      {pins.map((obj) => (
        <PinCard key={obj.pin.id} pin={obj} />
      ))}
    </div>
  )
}
