import { useState } from 'react'
import PinGrid from '../features/PinGrid'
import usePins from '../hooks/usePins'
import Header from '../features/Header'
import { useAuth } from '../context/authContext'

export default function PinPage({}) {
  const [query, setQuery] = useState('')
  const [avaUrl, setAvaUrl] = useState()
  const { pins, loading, error } = usePins(query)
  const { setAccessToken } = useAuth()

  function onAuthSuccess(data) {
    setAccessToken(data.token)
  }

  return (
    <div className='flex flex-col gap-4 min-h-dvh p-6'>
      <Header
        onAuthSuccess={onAuthSuccess}
        value={query}
        onChange={setQuery}
        avaUrl={avaUrl}
      />
      <PinGrid pins={pins} loading={loading} error={error} />
    </div>
  )
}
