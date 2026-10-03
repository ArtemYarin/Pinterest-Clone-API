import { useEffect, useId, useRef, useState } from 'react'
import Modal from '../Modal'
import { useAuth } from '../../context/authContext'

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const WEAK_PASSWORD =
  'This password is too easy to guess. Try a longer phrase or mix words, numbers and symbols.'

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

// Email + password form shared by login and sign up.
// `submit` posts the credentials and resolves to { token }.
// `statusMessages` maps HTTP statuses to user-facing errors.
export default function AuthModal({
  title,
  submitLabel,
  pendingLabel,
  passwordAutoComplete,
  submit,
  statusMessages = {},
  fallbackError,
  onClose,
}) {
  const { signIn } = useAuth()
  const id = useId()
  const emailRef = useRef(null)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [fieldErrors, setFieldErrors] = useState({})
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  // Runs after Modal's showModal(), which would otherwise focus the Close button.
  useEffect(() => emailRef.current.focus(), [])

  const handleSubmit = async (e) => {
    e.preventDefault()
    const errors = validate({ email, password })
    setFieldErrors(errors)
    setError('')
    if (Object.keys(errors).length) return

    setLoading(true)
    try {
      const data = await submit({ email: email.trim(), password_hash: password })
      onClose()
      signIn(data)
    } catch (err) {
      // 422 carries per-field validation errors: { Details: { password: 'weak password' } }
      const details = err.response?.data?.Details
      if (details) {
        setFieldErrors({
          email: details.email,
          password: details.password === 'weak password' ? WEAK_PASSWORD : details.password,
        })
      } else {
        const status = err.response?.status
        setError(statusMessages[status] || err.response?.data?.message || fallbackError)
      }
      setLoading(false)
    }
  }

  const inputBorder = (field) =>
    fieldErrors[field] ? 'border-red-400' : 'border-border hover:border-border-focus'

  return (
    <Modal onClose={onClose} labelledBy={`${id}-title`}>
      <form
        className='flex w-110 max-w-full flex-col items-center gap-8'
        onSubmit={handleSubmit}
        noValidate
      >
        <h2 id={`${id}-title`} className='font-display text-title font-bold'>
          {title}
        </h2>

        {error && (
          <p role='alert' className='w-full rounded-button bg-red-500/15 px-4 py-3 text-red-300'>
            {error}
          </p>
        )}

        <div className='w-full'>
          <label className='mb-1 block' htmlFor={`${id}-email`}>
            Email
          </label>
          <input
            ref={emailRef}
            className={`w-full rounded-button border px-4 py-3 ${inputBorder('email')}`}
            id={`${id}-email`}
            name='email'
            type='email'
            autoComplete='email'
            placeholder='you@example.com'
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            aria-invalid={!!fieldErrors.email}
            aria-describedby={fieldErrors.email ? `${id}-email-error` : undefined}
          />
          {fieldErrors.email && (
            <p id={`${id}-email-error`} className='mt-1 text-sm text-red-300'>
              {fieldErrors.email}
            </p>
          )}
        </div>

        <div className='w-full'>
          <label className='mb-1 block' htmlFor={`${id}-password`}>
            Password
          </label>
          <div className={`flex rounded-button border ${inputBorder('password')}`}>
            <input
              className='w-full bg-transparent px-4 py-3 outline-none'
              id={`${id}-password`}
              name='password'
              type={showPassword ? 'text' : 'password'}
              autoComplete={passwordAutoComplete}
              placeholder='••••••••'
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              aria-invalid={!!fieldErrors.password}
              aria-describedby={fieldErrors.password ? `${id}-password-error` : undefined}
            />
            <button
              className='rounded-button p-button py-2 hover:bg-card-focus'
              type='button'
              onClick={() => setShowPassword((s) => !s)}
              aria-pressed={showPassword}
            >
              {showPassword ? 'Hide' : 'Show'}
            </button>
          </div>
          {fieldErrors.password && (
            <p id={`${id}-password-error`} className='mt-1 text-sm text-red-300'>
              {fieldErrors.password}
            </p>
          )}
        </div>

        <button
          className='w-full rounded-button bg-accent p-button font-semibold text-bg hover:bg-accent-focus disabled:opacity-60'
          type='submit'
          disabled={loading}
        >
          {loading ? pendingLabel : submitLabel}
        </button>
      </form>
    </Modal>
  )
}
