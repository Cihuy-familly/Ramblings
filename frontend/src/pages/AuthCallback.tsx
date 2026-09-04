import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { CheckCircle2, XCircle } from 'lucide-react'
import { useAuth } from '../context/AuthContext'

export default function AuthCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setToken } = useAuth()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const token = searchParams.get('token')
    const errorParam = searchParams.get('error')

    if (errorParam) {
      setStatus('error')
      setError(decodeURIComponent(errorParam))
      return
    }

    if (!token) {
      setStatus('error')
      setError('No token received from authentication')
      return
    }

    // Store token and redirect to dashboard
    try {
      setToken(token)
      setStatus('success')

      // Redirect after a brief delay
      setTimeout(() => {
        navigate('/dashboard', { replace: true })
      }, 1500)
    } catch (err) {
      setStatus('error')
      setError('Failed to process authentication')
    }
  }, [searchParams, navigate, setToken])

  return (
    <div className="min-h-screen flex items-center justify-center px-4 pt-16">
      <div className="w-full max-w-sm">
        <div className="bg-white rounded-2xl shadow-lg border border-gray-200 p-8 text-center">
          {status === 'loading' && (
            <>
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand-600 mx-auto mb-4"></div>
              <h2 className="text-lg font-semibold text-gray-900">Signing you in...</h2>
              <p className="text-sm text-gray-500 mt-1">Please wait a moment</p>
            </>
          )}

          {status === 'success' && (
            <>
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-green-100 mb-4">
                <CheckCircle2 className="w-7 h-7 text-green-600" />
              </div>
              <h2 className="text-lg font-semibold text-gray-900">Signed in!</h2>
              <p className="text-sm text-gray-500 mt-1">Redirecting to dashboard...</p>
            </>
          )}

          {status === 'error' && (
            <>
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-red-100 mb-4">
                <XCircle className="w-7 h-7 text-red-600" />
              </div>
              <h2 className="text-lg font-semibold text-gray-900">Authentication failed</h2>
              <p className="text-sm text-red-600 mt-1">{error}</p>
              <button
                onClick={() => navigate('/login')}
                className="mt-6 px-4 py-2 text-sm font-medium text-white bg-brand-600 rounded-lg hover:bg-brand-700 transition-colors"
              >
                Back to login
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}