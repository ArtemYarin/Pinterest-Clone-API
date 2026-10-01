import icon from './../assets/default-user-icon.avif'

export default function Avatar({ url }) {
  const src = url ? url : icon
  const alt = url ? 'User avatar image' : 'Default avatar'

  return <img className='rounded-full size-10' src={src} alt={alt} />
}
