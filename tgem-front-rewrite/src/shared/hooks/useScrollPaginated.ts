import { useEffect } from "react"

// Wires window scroll → fetchNextPage when the user reaches the bottom.
// fetchNextPage from @tanstack/react-query is reference-stable, so the deps
// array stays clean.
export default function useScrollPaginated(fetchNextPage: () => unknown) {
  useEffect(() => {
    const handler = () => {
      if (
        window.innerHeight + document.documentElement.scrollTop
          !== document.documentElement.offsetHeight
      ) return
      fetchNextPage()
    }
    window.addEventListener("scroll", handler)
    return () => window.removeEventListener("scroll", handler)
  }, [fetchNextPage])
}
