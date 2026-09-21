import { useEffect, useState } from 'react'
import { getPins } from '../api/pins'
import PinCard from '../components/PinCard'

export default function PinGrid({ pins, loading, error }) {
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
