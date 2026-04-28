import { useState } from 'react'
import type { Geometry } from 'geojson'
import { DrawPanel } from './DrawPanel'
import { UploadPanel } from './UploadPanel'
import { PastePanel } from './PastePanel'

type Tab = 'draw' | 'upload' | 'paste'

interface InputMethodTabsProps {
  onGeometry: (g: Geometry) => void
}

export function InputMethodTabs({ onGeometry }: InputMethodTabsProps) {
  const [activeTab, setActiveTab] = useState<Tab>('draw')

  return (
    <div>
      <div className="input-tabs">
        {(['draw', 'upload', 'paste'] as Tab[]).map((tab) => (
          <button
            key={tab}
            className={`input-tab${activeTab === tab ? ' input-tab--active' : ''}`}
            onClick={() => setActiveTab(tab)}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>
      {activeTab === 'draw' && <DrawPanel />}
      {activeTab === 'upload' && <UploadPanel onGeometry={onGeometry} />}
      {activeTab === 'paste' && <PastePanel onGeometry={onGeometry} />}
    </div>
  )
}
