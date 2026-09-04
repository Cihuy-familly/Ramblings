import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '../context/AuthContext'
import Landing from '../pages/Landing'

describe('Landing page', () => {
  it('renders the hero heading', () => {
    render(
      <MemoryRouter>
        <AuthProvider>
          <Landing />
        </AuthProvider>
      </MemoryRouter>
    )

    expect(screen.getByText('Brambler')).toBeInTheDocument()
  })
})