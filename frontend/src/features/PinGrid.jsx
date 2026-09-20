import { useEffect, useState } from 'react'
import { getPins } from '../api/pins'
import PinCard from '../components/PinCard'

export default function PinGrid() {
  const [pins, setPins] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    getPins()
      .then(setPins)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p>Loading pins...</p>
  if (error) return <p>Error: {error}</p>

  return (
    <div className='columns-4 gap-8'>
      {pins.map((obj) => (
        <PinCard key={obj.pin.id} pin={obj} />
      ))}
    </div>
  )
}
