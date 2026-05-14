import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactNode, useState } from 'react'
import toast from 'react-hot-toast'

interface Props{
  children: ReactNode
}

export const ReactQueryProvider = ({children}: Props) => {
  const [queryClient] = useState(() => new QueryClient({
    queryCache: new QueryCache({
      onError(error) {
        toast.error((error as Error).message)
      },
    }),
    mutationCache: new MutationCache({
      onError(error) {
        toast.error((error as Error).message)
      },
    }),

    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false,
        networkMode: 'always',
        retry: false,
      },
      mutations: {
        networkMode: 'always',
        retry: false,
      }
    },
  }))
  
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}
