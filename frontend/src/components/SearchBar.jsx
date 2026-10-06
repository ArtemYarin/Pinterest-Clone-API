export default function SearchBar({ value, onChange, placeholder = 'Search' }) {
  return (
    <input
      id='search'
      type='search'
      aria-label='Search pins'
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className='w-full min-w-0 flex-1 text-big py-2 px-4 rounded-card bg-bg hover:bg-bg-focus outline-none focus-visible:ring-2 focus-visible:ring-accent'
    />
  )
}
