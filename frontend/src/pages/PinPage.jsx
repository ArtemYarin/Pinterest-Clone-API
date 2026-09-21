import { useState } from 'react'
import SearchBar from '../components/SearchBar'
import PinGrid from '../features/PinGrid'
import usePins from '../hooks/usePins'

export default function PinPage({}) {
  const [query, setQuery] = useState('')
  const { pins, loading, error } = usePins(query)

  return (
    <>
      <SearchBar value={query} onChange={setQuery} />
      <PinGrid pins={pins} loading={loading} error={error} />
    </>
  )
}
