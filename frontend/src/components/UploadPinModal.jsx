import { useEffect, useId, useRef, useState } from 'react'
import { useNavigate } from 'react-router'
import Modal from './Modal'
import { createPin, uploadPinImage, confirmPinUpload, deletePin } from '../api/pins'

// Mirrors the pin service limits checked on confirm.
const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
const MAX_IMAGE_BYTES = 10 << 20 // 10MB
const MAX_TITLE = 255
const MAX_DESCRIPTION = 1000

function validate({ file, title, description }) {
  const errors = {}
  if (!file) errors.image = 'Choose an image to upload.'
  else if (!ALLOWED_TYPES.includes(file.type))
    errors.image = 'Use a JPG, PNG, WebP or GIF image.'
  else if (file.size > MAX_IMAGE_BYTES) errors.image = 'Images must be 10MB or smaller.'
  if (!title.trim()) errors.title = 'Enter a title.'
  else if (title.trim().length > MAX_TITLE)
    errors.title = `Titles are at most ${MAX_TITLE} characters.`
  if (description.trim().length > MAX_DESCRIPTION)
    errors.description = `Descriptions are at most ${MAX_DESCRIPTION} characters.`
  return errors
}

// Creates a pin in three steps: post the metadata, PUT the image to the
// presigned upload url, then ask the backend to confirm the upload.
export default function UploadPinModal({ onClose }) {
  const id = useId()
  const navigate = useNavigate()
  const fileRef = useRef(null)
  const [file, setFile] = useState(null)
  const [preview, setPreview] = useState(null)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [fieldErrors, setFieldErrors] = useState({})
  const [error, setError] = useState('')
  // null | 'creating' | 'uploading' | 'confirming'
  const [step, setStep] = useState(null)
  const [progress, setProgress] = useState(0)

  // Runs after Modal's showModal(), which would otherwise focus the Close button.
  useEffect(() => fileRef.current.focus(), [])

  // Free the last preview url when the modal closes.
  useEffect(() => () => preview && URL.revokeObjectURL(preview), [preview])

  const handleFileChange = (e) => {
    const next = e.target.files[0] ?? null
    setFile(next)
    setPreview(next ? URL.createObjectURL(next) : null)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    const errors = validate({ file, title, description })
    setFieldErrors(errors)
    setError('')
    if (Object.keys(errors).length) return

    let pinId = null
    try {
      setStep('creating')
      const { pin, upload_url } = await createPin({
        title: title.trim(),
        ...(description.trim() && { description: description.trim() }),
      })
      pinId = pin.id

      setStep('uploading')
      setProgress(0)
      try {
        await uploadPinImage(upload_url, file, { onProgress: setProgress })
      } catch (err) {
        // The pin has no image; remove it now instead of waiting for the cleanup worker.
        deletePin(pinId).catch(() => {})
        throw err
      }

      setStep('confirming')
      await confirmPinUpload(pinId)

      onClose()
      navigate(`/${pinId}`)
    } catch (err) {
      const status = err.response?.status
      // 422 on create carries per-field errors: { Details: { Title: '...' } }
      const details = err.response?.data?.Details
      if (details) {
        setFieldErrors({ title: details.Title, description: details.Description })
      } else if (status === 422 && pinId) {
        setError('The image was rejected. Use a JPG, PNG, WebP or GIF up to 10MB.')
      } else if (!err.response && pinId) {
        setError('Image upload failed. Check your connection and try again.')
      } else {
        setError(err.response?.data?.message || 'Could not create the pin. Please try again.')
      }
      setStep(null)
    }
  }

  const inputBorder = (field) =>
    fieldErrors[field] ? 'border-red-400' : 'border-border hover:border-border-focus'

  const submitLabel = {
    creating: 'Creating…',
    uploading: `Uploading… ${progress}%`,
    confirming: 'Finishing…',
  }[step] ?? 'Create pin'

  return (
    <Modal onClose={onClose} labelledBy={`${id}-title`}>
      <form
        className='flex w-110 max-w-full flex-col items-center gap-6'
        onSubmit={handleSubmit}
        noValidate
      >
        <h2 id={`${id}-title`} className='font-display text-title font-bold'>
          Create pin
        </h2>

        {error && (
          <p role='alert' className='w-full rounded-button bg-red-500/15 px-4 py-3 text-red-300'>
            {error}
          </p>
        )}

        <div className='w-full'>
          <label className='mb-1 block' htmlFor={`${id}-image`}>
            Image
          </label>
          {preview && (
            <img
              className='mb-2 max-h-60 w-full rounded-button bg-bg-focus object-contain'
              src={preview}
              alt='Selected image preview'
            />
          )}
          <input
            ref={fileRef}
            className={`w-full rounded-button border px-4 py-3 file:mr-4 file:rounded-button file:border-0 file:bg-card-focus file:px-3 file:py-1 file:font-semibold file:text-text ${inputBorder('image')}`}
            id={`${id}-image`}
            name='image'
            type='file'
            accept={ALLOWED_TYPES.join(',')}
            onChange={handleFileChange}
            aria-invalid={!!fieldErrors.image}
            aria-describedby={fieldErrors.image ? `${id}-image-error` : undefined}
          />
          {fieldErrors.image && (
            <p id={`${id}-image-error`} className='mt-1 text-sm text-red-300'>
              {fieldErrors.image}
            </p>
          )}
        </div>

        <div className='w-full'>
          <label className='mb-1 block' htmlFor={`${id}-pin-title`}>
            Title
          </label>
          <input
            className={`w-full rounded-button border px-4 py-3 ${inputBorder('title')}`}
            id={`${id}-pin-title`}
            name='title'
            type='text'
            maxLength={MAX_TITLE}
            placeholder='Add a title'
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            aria-invalid={!!fieldErrors.title}
            aria-describedby={fieldErrors.title ? `${id}-pin-title-error` : undefined}
          />
          {fieldErrors.title && (
            <p id={`${id}-pin-title-error`} className='mt-1 text-sm text-red-300'>
              {fieldErrors.title}
            </p>
          )}
        </div>

        <div className='w-full'>
          <label className='mb-1 block' htmlFor={`${id}-description`}>
            Description <span className='opacity-60'>(optional)</span>
          </label>
          <textarea
            className={`w-full resize-y rounded-button border bg-transparent px-4 py-3 ${inputBorder('description')}`}
            id={`${id}-description`}
            name='description'
            rows={3}
            maxLength={MAX_DESCRIPTION}
            placeholder='Tell everyone what your pin is about'
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            aria-invalid={!!fieldErrors.description}
            aria-describedby={fieldErrors.description ? `${id}-description-error` : undefined}
          />
          {fieldErrors.description && (
            <p id={`${id}-description-error`} className='mt-1 text-sm text-red-300'>
              {fieldErrors.description}
            </p>
          )}
        </div>

        <button
          className='w-full rounded-button bg-accent p-button font-semibold text-bg hover:bg-accent-focus disabled:opacity-60'
          type='submit'
          disabled={!!step}
        >
          {submitLabel}
        </button>
      </form>
    </Modal>
  )
}
