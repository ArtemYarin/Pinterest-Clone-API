import PinCard from '../components/PinCard'

export default function PinGrid({ pins, loading, error }) {
  if (loading) return <p>Loading pins...</p>
  if (error) return <p>Error: {error}</p>

  console.log(pins)

  if (pins === null) {
    return (
      <div className='mt-16 flex flex-1 justify-center items-center'>
        <p className='font-display text-5xl opacity-75'>No pins</p>
      </div>
    )
  } else {
    return (
      <div className='mt-16 flex-1 columns-4 gap-6'>
        {pins.map((obj) => (
          <PinCard key={obj.pin.id} pin={obj} />
        ))}
      </div>
    )
  }
}
