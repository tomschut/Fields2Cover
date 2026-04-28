import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PipelineErrorBanner } from './PipelineErrorBanner'

describe('PipelineErrorBanner', () => {
  it('renders error message when error is non-null', () => {
    render(<PipelineErrorBanner error="Something went wrong" />)
    expect(screen.getByText('Something went wrong')).toBeDefined()
  })

  it('renders nothing when error is null', () => {
    const { container } = render(<PipelineErrorBanner error={null} />)
    expect(container.firstChild).toBeNull()
  })
})
