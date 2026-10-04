import { useState } from 'react'
import PinGrid from '../features/PinGrid'
import usePins from '../hooks/usePins'
import useDebouncedValue from '../hooks/useDebouncedValue'
import Header from '../features/Header'

export default function PinsPage() {
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebouncedValue(query.trim(), 300)
  const { pins, loading, error } = usePins(debouncedQuery)

  return (
    <div className='flex min-h-dvh flex-col gap-4 px-6 pb-6'>
      <Header value={query} onChange={setQuery} />
      <PinGrid pins={pins} loading={loading} error={error} />
    </div>
  )
}
