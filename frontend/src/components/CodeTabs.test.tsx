import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { CodeTabs } from './CodeTabs'

const CURL_CONTENT = 'curl -X POST http://localhost:8080/pipeline/plan-coverage'
const JS_CONTENT = 'fetch("/pipeline/plan-coverage"'
const GO_CONTENT = 'http.NewRequest("POST", "/pipeline/plan-coverage"'

describe('CodeTabs', () => {
  it('renders all three tab buttons', () => {
    render(<CodeTabs curl={CURL_CONTENT} js={JS_CONTENT} go={GO_CONTENT} />)
    expect(screen.getByText('curl')).toBeDefined()
    expect(screen.getByText('JavaScript')).toBeDefined()
    expect(screen.getByText('Go')).toBeDefined()
  })

  it('curl tab is active by default', () => {
    render(<CodeTabs curl={CURL_CONTENT} js={JS_CONTENT} go={GO_CONTENT} />)
    const curlBtn = screen.getByText('curl')
    expect(curlBtn.className).toContain('code-tab--active')
    const jsBtn = screen.getByText('JavaScript')
    expect(jsBtn.className).not.toContain('code-tab--active')
  })

  it('renders curl content in pre.code-block by default', () => {
    render(<CodeTabs curl={CURL_CONTENT} js={JS_CONTENT} go={GO_CONTENT} />)
    expect(screen.getByText(CURL_CONTENT)).toBeDefined()
  })

  it('switches content to JS when JavaScript tab is clicked', () => {
    render(<CodeTabs curl={CURL_CONTENT} js={JS_CONTENT} go={GO_CONTENT} />)
    fireEvent.click(screen.getByText('JavaScript'))
    expect(screen.queryByText(CURL_CONTENT)).toBeNull()
    expect(screen.getByText(JS_CONTENT)).toBeDefined()
  })
})
