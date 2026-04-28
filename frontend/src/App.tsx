import { Routes, Route, NavLink } from 'react-router-dom'
import MapPage from './pages/MapPage'
import GuidePage from './pages/GuidePage'
import ApiDocsPage from './pages/ApiDocsPage'
function App() {
  return (
    <div className="app">
      <header className="header">
        <div className="header-inner">
          <span className="wordmark">Fields2Cover</span>
          <nav>
            <NavLink to="/" end className={({ isActive }) => isActive ? 'nav-link--active' : undefined}>
              Map Demo
            </NavLink>
            <NavLink to="/guide" className={({ isActive }) => isActive ? 'nav-link--active' : undefined}>
              Integration Guide
            </NavLink>
            <NavLink to="/api-docs" className={({ isActive }) => isActive ? 'nav-link--active' : undefined}>
              API Docs
            </NavLink>
          </nav>
        </div>
      </header>
      <Routes>
        <Route path="/" element={<MapPage />} />
        <Route path="/guide" element={<GuidePage />} />
        <Route path="/api-docs" element={<ApiDocsPage />} />
      </Routes>
    </div>
  )
}

export default App
