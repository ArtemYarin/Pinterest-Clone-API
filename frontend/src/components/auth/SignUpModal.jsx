import { useState } from 'react'
import { signup } from '../../api/signup'

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function validate({ email, password }) {
  const errors = {}
  if (!email.trim()) errors.email = 'Enter your email address.'
  else if (!EMAIL_PATTERN.test(email.trim()))
    errors.email = 'Enter an email in the format name@example.com.'
  if (!password) errors.password = 'Enter your password.'
  else if (password.length < 8)
    errors.password = 'Passwords are at least 8 characters.'
  return errors
}

export default function SignUpModal({ onClose, onSuccess }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    signup({ email: email, password_hash: password })
      .then((res) => {
        onClose()
        onSuccess(res)
      })
      .catch((err) => {
        const msg = err.response?.data?.message || 'Sign up failed'
        setError(msg)
      })
      .finally(() => {
        setLoading(false)
      })
  }

  return (
    <div
      className='fixed inset-0 z-50 flex items-center justify-center p-0 bg-bg/40 backdrop-blur-[2px]'
      onClick={(e) => {
        // target is the element that was actually clicked.
        // currentTarget is the element the handler is attached to.
        if (e.target === e.currentTarget) onClose()
      }}
    >
      {/* Card */}
      <div className='flex flex-col items-center justify-center p-card pt-8 rounded-card bg-card'>
        {/* Close button */}
        <button
          type='button'
          onClick={onClose}
          aria-label='Close'
          className='relative -top-4 -right-6 self-end rounded-full p-button hover:bg-card-focus'
        >
          Close
        </button>

        <form
          className='flex flex-col items-center justify-center gap-8 mb-4 w-110'
          onSubmit={handleSubmit}
        >
          <h2 className='font-display font-extrabold text-title'>Sign up</h2>
          <div className='w-full'>
            <label className='block mb-1' htmlFor='email'>
              Email
            </label>
            <input
              className='w-full px-4 py-3 border border-border hover:border-focus rounded-button'
              id='email'
              name='email'
              type='email'
              autoComplete='email'
              placeholder='you@example.com'
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <div className='w-full'>
            <label className='block mb-1' htmlFor='password'>
              Password
            </label>
            <div className='flex justify-between gap-0 border border-border rounded-button hover:border-focus'>
              <input
                className='w-full bg-transparent outline-none px-4 py-3'
                id='password'
                name='password'
                type={showPassword ? 'text' : 'password'}
                autoComplete='current-password'
                placeholder='••••••••'
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <button
                className='rounded-button p-button py-2 hover:bg-card-focus'
                type='button'
                onClick={() => setShowPassword((s) => !s)}
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
          </div>
          <button
            className='w-full p-button rounded-button bg-accent hover:bg-accent-focus'
            type='submit'
            disabled={loading}
          >
            {loading ? 'Signing up' : 'Sign up'}
          </button>
        </form>
      </div>
    </div>
  )
}
