import { vi, describe, it, expect, beforeEach } from 'vitest'
import { planCoverage, planSwaths, pickBestVariant } from './pipeline'
import type { PlanCoverageRequest } from './pipeline'

vi.mock('ky', () => ({
  default: {
    post: vi.fn(),
  },
  HTTPError: class HTTPError extends Error {
    response: { status: number }
    constructor(response: { status: number }) {
      super('HTTPError')
      this.response = response
    }
  },
}))

import ky, { HTTPError } from 'ky'

const mockPost = vi.mocked(ky.post)

const sampleReq: PlanCoverageRequest = {
  geojson: { type: 'FeatureCollection', features: [] },
  robot: { width_m: 3, cov_width_m: 3 },
  headland_width_m: 6,
  swath_angle_rad: 0,
  swath_width_m: 3,
  sort_algorithm: 'BOUSTROPHEDON',
  turning_algorithm: 'DUBINS',
}

const sampleReqWithVariant: PlanCoverageRequest = {
  geojson: { type: 'FeatureCollection', features: [] },
  robot: { width_m: 3, cov_width_m: 3 },
  headland_width_m: 6,
  swath_angle_rad: 0,
  swath_width_m: 3,
  sort_algorithm: 'BOUSTROPHEDON',
  turning_algorithm: 'DUBINS',
  sort_variant: 1,
}

describe('planSwaths', () => {
  beforeEach(() => {
    const mockJsonFn = vi.fn().mockResolvedValue({ field_with_headlands: {}, sorted_swaths: { items: [] }, journey: {} })
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
  })

  it('calls /pipeline/plan-swaths endpoint', async () => {
    await planSwaths(sampleReqWithVariant)
    expect(mockPost).toHaveBeenCalledWith('/pipeline/plan-swaths', { json: sampleReqWithVariant })
  })

  it('calls /pipeline/plan-swaths with request without sort_variant', async () => {
    await planSwaths(sampleReq)
    expect(mockPost).toHaveBeenCalledWith('/pipeline/plan-swaths', { json: sampleReq })
  })

  it('throws "Invalid parameters" message on HTTPError 400', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const mockJsonFn = vi.fn().mockRejectedValue(new (HTTPError as any)({ status: 400 }))
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
    await expect(planSwaths(sampleReqWithVariant)).rejects.toThrow('Invalid parameters.')
  })

  it('throws "Server error" message on HTTPError 500', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const mockJsonFn = vi.fn().mockRejectedValue(new (HTTPError as any)({ status: 500 }))
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
    await expect(planSwaths(sampleReqWithVariant)).rejects.toThrow('Server error.')
  })

  it('throws "Could not reach the server" on network error', async () => {
    const mockJsonFn = vi.fn().mockRejectedValue(new Error('network failure'))
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
    await expect(planSwaths(sampleReqWithVariant)).rejects.toThrow('Could not reach the server.')
  })
})

describe('pickBestVariant', () => {
  it('returns variant with closest firstSwathStart to startPoint', () => {
    const startPoint: [number, number] = [52.0, 5.0] // [lat, lng]
    const attempts = [
      { variant: 0, firstSwathStart: [5.1, 52.1] as [number, number] }, // [lng, lat] far
      { variant: 1, firstSwathStart: [5.01, 52.01] as [number, number] }, // closest
      { variant: 2, firstSwathStart: [5.2, 52.2] as [number, number] },
      { variant: 3, firstSwathStart: [5.3, 52.3] as [number, number] },
    ]
    expect(pickBestVariant(startPoint, attempts)).toBe(1)
  })

  it('returns 0 when all firstSwathStart values are null', () => {
    const startPoint: [number, number] = [52.0, 5.0]
    const attempts = [
      { variant: 0, firstSwathStart: null },
      { variant: 1, firstSwathStart: null },
      { variant: 2, firstSwathStart: null },
      { variant: 3, firstSwathStart: null },
    ]
    expect(pickBestVariant(startPoint, attempts)).toBe(0)
  })

  it('ignores null entries and picks best from remaining', () => {
    const startPoint: [number, number] = [52.0, 5.0]
    const attempts = [
      { variant: 0, firstSwathStart: null },
      { variant: 2, firstSwathStart: [5.0, 52.0] as [number, number] }, // exact match
    ]
    expect(pickBestVariant(startPoint, attempts)).toBe(2)
  })
})

describe('planCoverage', () => {
  beforeEach(() => {
    const mockJsonFn = vi.fn().mockResolvedValue({ path: {}, intermediates: {}, journey: {} })
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
  })

  it('resolves with response on 200', async () => {
    const result = await planCoverage(sampleReq)
    expect(result).toBeDefined()
    expect(mockPost).toHaveBeenCalledWith('/pipeline/plan-coverage', { json: sampleReq })
  })

  it('throws "Invalid parameters" message on HTTPError 400', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const mockJsonFn = vi.fn().mockRejectedValue(new (HTTPError as any)({ status: 400 }))
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
    await expect(planCoverage(sampleReq)).rejects.toThrow('Invalid parameters.')
  })

  it('throws "Server error" message on HTTPError 500', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const mockJsonFn = vi.fn().mockRejectedValue(new (HTTPError as any)({ status: 500 }))
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
    await expect(planCoverage(sampleReq)).rejects.toThrow('Server error.')
  })

  it('throws "Could not reach the server" on network error', async () => {
    const mockJsonFn = vi.fn().mockRejectedValue(new Error('network failure'))
    mockPost.mockReturnValue({ json: mockJsonFn } as unknown as ReturnType<typeof ky.post>)
    await expect(planCoverage(sampleReq)).rejects.toThrow('Could not reach the server.')
  })

  it('passes sort_start_point through to the request body', async () => {
    await planCoverage({ ...sampleReq, sort_start_point: [5.1, 52.2] })
    expect(mockPost).toHaveBeenCalledWith(
      '/pipeline/plan-coverage',
      { json: { ...sampleReq, sort_start_point: [5.1, 52.2] } }
    )
  })
})
