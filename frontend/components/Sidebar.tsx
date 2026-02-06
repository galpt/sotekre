import React from 'react'

type SidebarProps = {
    activeMenu?: string
    onMenuClick?: (menu: string) => void
}

export default function Sidebar({ activeMenu = 'Menus', onMenuClick }: SidebarProps) {
    const menuItems = [
        { id: 'systems', icon: '📁', label: 'Systems' },
        { id: 'systemCode', icon: '💻', label: 'System Code' },
        { id: 'properties', icon: '⚙️', label: 'Properties' },
        { id: 'menus', icon: '☰', label: 'Menus' },
        { id: 'apiList', icon: '📋', label: 'API List' },
        { id: 'usersGroup', icon: '👥', label: 'Users & Group' },
        { id: 'competition', icon: '🏆', label: 'Competition' },
    ]

    return (
        <div className="w-[188px] h-screen bg-[#0D47A1] text-white flex flex-col fixed left-0 top-0">
            {/* Logo/Brand Section */}
            <div className="p-4 flex items-center justify-between border-b border-blue-700">
                <div className="flex items-center gap-2">
                    <div className="w-8 h-8 bg-white rounded flex items-center justify-center">
                        <div className="grid grid-cols-2 gap-0.5">
                            <div className="w-1.5 h-1.5 bg-[#0D47A1] rounded-sm"></div>
                            <div className="w-1.5 h-1.5 bg-[#0D47A1] rounded-sm"></div>
                            <div className="w-1.5 h-1.5 bg-[#0D47A1] rounded-sm"></div>
                            <div className="w-1.5 h-1.5 bg-[#0D47A1] rounded-sm"></div>
                        </div>
                    </div>
                    <div className="flex flex-col">
                        <span className="text-xs font-semibold leading-tight">Solusi</span>
                        <span className="text-xs font-semibold leading-tight">Teknologi</span>
                        <span className="text-xs font-semibold leading-tight">Kreatif</span>
                    </div>
                </div>
                <button className="text-white">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <line x1="3" y1="12" x2="21" y2="12"></line>
                        <line x1="3" y1="6" x2="21" y2="6"></line>
                        <line x1="3" y1="18" x2="21" y2="18"></line>
                    </svg>
                </button>
            </div>

            {/* Navigation Menu */}
            <nav className="flex-1 py-4">
                {menuItems.map((item) => {
                    const isActive = item.label === activeMenu
                    return (
                        <button
                            key={item.id}
                            onClick={() => onMenuClick?.(item.label)}
                            className={`w-full px-4 py-3 flex items-center gap-3 text-left transition-colors ${isActive
                                    ? 'bg-white text-[#0D47A1] font-medium'
                                    : 'text-white hover:bg-blue-800'
                                }`}
                        >
                            <span className={`text-base ${isActive ? '' : 'opacity-80'}`}>
                                {item.icon}
                            </span>
                            <span className="text-sm">{item.label}</span>
                        </button>
                    )
                })}
            </nav>
        </div>
    )
}
