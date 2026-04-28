import { useState } from 'react'

type Tab = 'curl' | 'js' | 'go'

interface CodeTabsProps {
  curl: string
  js: string
  go: string
}

export function CodeTabs({ curl, js, go }: CodeTabsProps) {
  const [activeTab, setActiveTab] = useState<Tab>('curl')

  const labels: Record<Tab, string> = {
    curl: 'curl',
    js: 'JavaScript',
    go: 'Go',
  }

  const content: Record<Tab, string> = { curl, js, go }

  return (
    <>
      <div className="code-tabs">
        {(['curl', 'js', 'go'] as Tab[]).map((tab) => (
          <button
            key={tab}
            className={`code-tab${activeTab === tab ? ' code-tab--active' : ''}`}
            onClick={() => setActiveTab(tab)}
          >
            {labels[tab]}
          </button>
        ))}
      </div>
      <pre className="code-block"><code>{content[activeTab]}</code></pre>
    </>
  )
}
