import React from 'react'
import ReactDOM from 'react-dom/client'
import App from "@app/App"
import "@app/index.css"
import { ReactQueryProvider } from "@app/providers/ReactQueryProvider"
import { BrowserRouter } from 'react-router-dom'
import AuthProvider from "@app/providers/AuthProvider"

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <ReactQueryProvider>
        <AuthProvider>
          <App />
        </AuthProvider>
      </ReactQueryProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
