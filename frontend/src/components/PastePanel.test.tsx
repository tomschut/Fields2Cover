import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PastePanel } from './PastePanel'

describe('PastePanel', () => {
  it('renders textarea and Load field button', () => {
    render(<PastePanel onGeometry={vi.fn()} />)
    expect(screen.getByRole('textbox')).toBeDefined()
    expect(screen.getByText('Load field')).toBeDefined()
  })

  it('shows error when clicking Load with empty textarea', async () => {
    render(<PastePanel onGeometry={vi.fn()} />)
    await userEvent.click(screen.getByText('Load field'))
    expect(screen.getByText('Paste GeoJSON before clicking Load.')).toBeDefined()
  })

  it('calls onGeometry with valid GeoJSON paste', async () => {
    const onGeometry = vi.fn()
    render(<PastePanel onGeometry={onGeometry} />)

    const validGeoJSON = JSON.stringify({
      type: 'Polygon',
      coordinates: [[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]],
    })

    // Use fireEvent.change to avoid userEvent.type interpreting { as keyboard modifier
    fireEvent.change(screen.getByRole('textbox'), { target: { value: validGeoJSON } })
    await userEvent.click(screen.getByText('Load field'))

    expect(onGeometry).toHaveBeenCalledOnce()
    expect(onGeometry.mock.calls[0][0].type).toBe('Polygon')
  })

  it('shows error for invalid JSON paste', async () => {
    render(<PastePanel onGeometry={vi.fn()} />)

    await userEvent.type(screen.getByRole('textbox'), 'not json at all')
    await userEvent.click(screen.getByText('Load field'))

    expect(screen.getByText(/Invalid GeoJSON/)).toBeDefined()
  })

  it('shows error for valid JSON but unsupported GeoJSON type', async () => {
    render(<PastePanel onGeometry={vi.fn()} />)

    const pointGeoJSON = JSON.stringify({ type: 'Point', coordinates: [5, 52] })
    // Use fireEvent.change to avoid userEvent.type interpreting { as keyboard modifier
    fireEvent.change(screen.getByRole('textbox'), { target: { value: pointGeoJSON } })
    await userEvent.click(screen.getByText('Load field'))

    expect(screen.getByText(/Invalid GeoJSON/)).toBeDefined()
  })

  it('clears error on successful paste after previous error', async () => {
    const onGeometry = vi.fn()
    render(<PastePanel onGeometry={onGeometry} />)

    // First: trigger error
    await userEvent.click(screen.getByText('Load field'))
    expect(screen.getByText('Paste GeoJSON before clicking Load.')).toBeDefined()

    // Second: paste valid GeoJSON
    const validGeoJSON = JSON.stringify({
      type: 'Polygon',
      coordinates: [[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]],
    })
    // Use fireEvent.change to avoid userEvent.type interpreting { as keyboard modifier
    fireEvent.change(screen.getByRole('textbox'), { target: { value: validGeoJSON } })
    await userEvent.click(screen.getByText('Load field'))

    expect(onGeometry).toHaveBeenCalledOnce()
    // Error should be cleared
    expect(screen.queryByText('Paste GeoJSON before clicking Load.')).toBeNull()
  })
})
