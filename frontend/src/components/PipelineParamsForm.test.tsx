import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import type { Geometry } from 'geojson'
import { PipelineParamsForm } from './PipelineParamsForm'

const mockGeometry: Geometry = { type: 'Polygon', coordinates: [[[0, 0], [1, 0], [1, 1], [0, 0]]] }

describe('PipelineParamsForm', () => {
  it('renders all 6 form fields', () => {
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={vi.fn()} loading={false} />)
    expect(screen.getByLabelText('Robot width (m)')).toBeDefined()
    expect(screen.getByLabelText('Coverage width (m)')).toBeDefined()
    expect(screen.getByLabelText('Speed (m/s)')).toBeDefined()
    expect(screen.getByLabelText('Headland width (m)')).toBeDefined()
    expect(screen.getByLabelText('Sort algorithm')).toBeDefined()
    expect(screen.getByLabelText('Turning algorithm')).toBeDefined()
  })

  it('robot width input default value is "3"', () => {
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={vi.fn()} loading={false} />)
    const input = screen.getByLabelText('Robot width (m)') as HTMLInputElement
    expect(input.value).toBe('3.0')
  })

  it('coverage width input default value is "3"', () => {
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={vi.fn()} loading={false} />)
    const input = screen.getByLabelText('Coverage width (m)') as HTMLInputElement
    expect(input.value).toBe('3.0')
  })

  it('headland width input default value is "6"', () => {
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={vi.fn()} loading={false} />)
    const input = screen.getByLabelText('Headland width (m)') as HTMLInputElement
    expect(input.value).toBe('6.0')
  })

  it('Run button is disabled when fieldGeometry prop is null', () => {
    render(<PipelineParamsForm fieldGeometry={null} onRun={vi.fn()} loading={false} />)
    const btn = screen.getByText('Run pipeline') as HTMLButtonElement
    expect(btn.disabled).toBe(true)
  })

  it('blur on empty robot width input shows error "Required"', () => {
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={vi.fn()} loading={false} />)
    const input = screen.getByLabelText('Robot width (m)')
    fireEvent.change(input, { target: { value: '' } })
    fireEvent.blur(input)
    expect(screen.getByText('Required')).toBeDefined()
  })

  it('blur on robot width input with value "0" shows error "Must be greater than 0"', () => {
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={vi.fn()} loading={false} />)
    const input = screen.getByLabelText('Robot width (m)')
    fireEvent.change(input, { target: { value: '0' } })
    fireEvent.blur(input)
    expect(screen.getByText('Must be greater than 0')).toBeDefined()
  })

  it('Run button calls onRun with sort_variant 0 (overridden by MapPage.handleRunPipeline at runtime)', () => {
    const onRun = vi.fn()
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={onRun} loading={false} />)
    const btn = screen.getByText('Run pipeline')
    fireEvent.click(btn)
    expect(onRun).toHaveBeenCalledOnce()
    expect(onRun.mock.calls[0][0]).toMatchObject({ sort_variant: 0 })
  })

  it('Run button calls onRun with correct PlanCoverageRequest shape when all fields valid and fieldGeometry non-null', () => {
    const onRun = vi.fn()
    render(<PipelineParamsForm fieldGeometry={mockGeometry} onRun={onRun} loading={false} />)
    const btn = screen.getByText('Run pipeline')
    fireEvent.click(btn)
    expect(onRun).toHaveBeenCalledOnce()
    expect(onRun.mock.calls[0][0]).toMatchObject({
      geojson: expect.objectContaining({ type: 'FeatureCollection' }),
      robot: expect.objectContaining({ width_m: 3, cov_width_m: 3 }),
    })
  })
})
