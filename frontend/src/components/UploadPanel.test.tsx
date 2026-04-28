import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { UploadPanel } from './UploadPanel'

describe('UploadPanel', () => {
  it('renders the drop zone with idle text', () => {
    render(<UploadPanel onGeometry={vi.fn()} />)
    expect(screen.getByText('Drop a .geojson file here, or click to browse')).toBeDefined()
  })

  it('has a hidden file input with .geojson accept', () => {
    render(<UploadPanel onGeometry={vi.fn()} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    expect(input).toBeDefined()
    expect(input.accept).toBe('.geojson,application/geo+json')
    expect(input.style.display).toBe('none')
  })

  it('calls onGeometry with valid GeoJSON file content', async () => {
    const onGeometry = vi.fn()
    render(<UploadPanel onGeometry={onGeometry} />)

    const validGeoJSON = JSON.stringify({
      type: 'Polygon',
      coordinates: [[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]],
    })

    const file = new File([validGeoJSON], 'field.geojson', { type: 'application/geo+json' })
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await userEvent.upload(input, file)

    // FileReader is async -- wait for callback
    await vi.waitFor(() => {
      expect(onGeometry).toHaveBeenCalledOnce()
    })
    expect(onGeometry.mock.calls[0][0].type).toBe('Polygon')
  })

  it('shows error for invalid JSON file', async () => {
    render(<UploadPanel onGeometry={vi.fn()} />)

    const file = new File(['not json'], 'bad.geojson', { type: 'application/geo+json' })
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await userEvent.upload(input, file)

    await vi.waitFor(() => {
      expect(screen.getByText('Could not read this file. Check that it is valid GeoJSON.')).toBeDefined()
    })
  })
})
