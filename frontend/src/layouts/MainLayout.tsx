import { Link, Outlet } from "react-router-dom";
import './MainLayout.css'

function MainLayout() {
    return (
        <div className="layout-container">
            {/* top header */}
            <header className="layout-header">
                <h1 className="layout-header-title">D-Come</h1>
            </header>

            {/* main body */}
            <div className="layout-body">
                {/* left sidebar */}
                <nav className="layout-sidebar">
                    <ul className="sidebar-nav-list">
                        <li className="sidebar-nav-item">
                            <Link to="/ah-stock" className="sidebar-nav-link">AH-Stock module</Link>
                        </li>
                        <li className="sidebar-nav-item">
                            <Link to="/custom" className="sidebar-nav-link">Custom module</Link>
                        </li>
                        <li className="sidebar-nav-item">
                            <Link to="/alarm" className="sidebar-nav-link">Alarm module</Link>
                        </li>
                    </ul>
                </nav>
                
                {/* right content area */}
                <main className="layout-content-area">
                    <Outlet />
                </main>
            </div>
        </div>
    )
}

export default MainLayout