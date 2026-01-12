/**
 * Search highlighting utilities with JSX support
 */

/**
 * Highlight matching text in a string
 */
export function highlightMatches(text: string, query: string): React.ReactNode {
  if (!query.trim()) return text

  const searchTerms = query.toLowerCase().trim().split(/\s+/)

  // Build a regex pattern that matches any of the search terms (case insensitive)
  const pattern = searchTerms.map(term =>
    term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // Escape special chars
  ).join('|')

  const regex = new RegExp(`(${pattern})`, 'gi')
  const parts = text.split(regex)

  return parts.map((part, index) => {
    // Check if this part matches any search term
    const isMatch = searchTerms.some(term =>
      part.toLowerCase() === term.toLowerCase()
    )

    if (isMatch) {
      return (
        <mark key={index} className="bg-yellow-200 dark:bg-yellow-800">
          {part}
        </mark>
      )
    }

    return part
  })
}
