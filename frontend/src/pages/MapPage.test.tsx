import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from '../App'

// Mock react-leaflet to avoid jsdom limitations with Leaflet DOM manipulation
const mockMap = {
  pm: { addControls: vi.fn(), removeControls: vi.fn() },
  on: vi.fn(),
  off: vi.fn(),
  removeLayer: vi.fn(),
  getZoom: vi.fn().mockReturnValue(8),
  getBounds: vi.fn().mockReturnValue({
    getWest: () => 3.3,
    getSouth: () => 50.5,
    getEast: () => 7.3,
    getNorth: () => 53.7,
  }),
}
vi.mock('react-leaflet', () => ({
  MapContainer: ({ children }: { children: React.ReactNode }) => <div data-testid="map-container">{children}</div>,
  TileLayer: () => <div data-testid="tile-layer" />,
  GeoJSON: () => <div data-testid="geojson-layer" />,
  Polyline: () => <div data-testid="polyline" />,
  CircleMarker: () => <div data-testid="circle-marker" />,
  useMap: () => mockMap,
  useMapEvents: (_handlers: unknown) => mockMap,
}))

vi.mock('../api/pipeline', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/pipeline')>()
  return {
    ...actual,
    planCoverage: vi.fn().mockResolvedValue({
      path: { states: [] },
      intermediates: { field_with_headlands: null, sorted_swaths: null, route: null },
      journey: {},
    }),
    planSwaths: vi.fn().mockResolvedValue({
      sorted_swaths: { items: [{ path: { coordinates: [[5.659, 51.997]] } }] },
      field_with_headlands: {},
      journey: {},
    }),
  }
})

vi.mock('@geoman-io/leaflet-geoman-free', () => ({}))

// Mock swagger-ui-react to avoid jsdom limitations (same pattern as react-leaflet mock)
vi.mock('swagger-ui-react', () => ({
  default: () => <div data-testid="swagger-ui" />,
}))

describe('MapPage', () => {
  it('renders the map page at / route', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByTestId('map-container')).toBeDefined()
  })

  it('renders navigation links with title-case labels', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByText('Map Demo')).toBeDefined()
    expect(screen.getByText('Integration Guide')).toBeDefined()
    expect(screen.getByText('API Docs')).toBeDefined()
  })

  it('renders GuidePage at /guide route (not "Coming soon")', () => {
    render(
      <MemoryRouter initialEntries={['/guide']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getAllByText('Integration Guide').length).toBeGreaterThan(0)
  })

  it('renders ApiDocsPage at /api-docs route', () => {
    render(
      <MemoryRouter initialEntries={['/api-docs']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByTestId('swagger-ui')).toBeDefined()
  })

  it('shows "No field defined" initially', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByText('No field defined')).toBeDefined()
  })

  it('renders "Set start" button on the map page', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByTitle('Set start point')).toBeDefined()
  })
})
