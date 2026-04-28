import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ResultLayerToggles } from './ResultLayerToggles'
import type { LayerVisibility } from './ResultLayerToggles'

const allOn: LayerVisibility = { headlands: true, swaths: true, route: true, path: true }

describe('ResultLayerToggles', () => {
  it('renders four checkboxes', () => {
    render(<ResultLayerToggles layerVisibility={allOn} onVisibilityChange={() => {}} />)
    const checkboxes = screen.getAllByRole('checkbox')
    expect(checkboxes).toHaveLength(4)
  })

  it('all checkboxes are checked by default when layerVisibility is all-true', () => {
    render(<ResultLayerToggles layerVisibility={allOn} onVisibilityChange={() => {}} />)
    const checkboxes = screen.getAllByRole('checkbox') as HTMLInputElement[]
    checkboxes.forEach(cb => expect(cb.checked).toBe(true))
  })

  it('calls onVisibilityChange with headlands false when Headlands unchecked', () => {
    const onChange = vi.fn()
    render(<ResultLayerToggles layerVisibility={allOn} onVisibilityChange={onChange} />)
    const headlandsCheckbox = screen.getByRole('checkbox', { name: /headlands/i })
    fireEvent.click(headlandsCheckbox)
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ headlands: false }))
  })

  it('renders all four layer labels', () => {
    render(<ResultLayerToggles layerVisibility={allOn} onVisibilityChange={() => {}} />)
    expect(screen.getByText(/headlands/i)).toBeDefined()
    expect(screen.getByText(/swaths/i)).toBeDefined()
    expect(screen.getByText(/route/i)).toBeDefined()
    expect(screen.getByText(/path/i)).toBeDefined()
  })
})
