import { useState } from 'react'
import SignUpModal from './SignUpModal'

export default function SignupButton({ onSuccess }) {
  const [showModal, setShowModal] = useState(false)
  return (
    <>
      <button
        className='bg-accent text-bg font-semibold whitespace-nowrap p-button rounded-button hover:bg-accent-focus'
        type='button'
        onClick={() => setShowModal(true)}
      >
        Sign Up
      </button>
      {showModal && (
        <SignUpModal
          onClose={() => setShowModal(false)}
          onSuccess={onSuccess}
        />
      )}
    </>
  )
}
