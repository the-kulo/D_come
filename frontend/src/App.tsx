import { Routes, Route, Navigate } from 'react-router-dom'
import MainLayout from './layouts/MainLayout'
import AHStock from './pages/AH_Stock module'
import Custom from './pages/Custom module'    
import Alarm from './pages/Alarm module'      

function App() {
  return (
    <Routes>
      <Route path="/" element={<MainLayout />}>
        <Route index element={<Navigate to="/ah-stock" replace />} />
        <Route path="ah-stock" element={<AHStock />} />
        <Route path="custom" element={<Custom />} />
        <Route path="alarm" element={<Alarm />} />
      </Route>
    </Routes>
  )
}

export default App