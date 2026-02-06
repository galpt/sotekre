import React from 'react'

type SidebarProps = {
    activeMenu?: string
    onMenuClick?: (menu: string) => void
}

// SVG Icons
const FolderOpenIcon = () => (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
        <path d="M1.5 2A1.5 1.5 0 000 3.5v9A1.5 1.5 0 001.5 14h13a1.5 1.5 0 001.5-1.5v-7A1.5 1.5 0 0014.5 4H9.414L7.707 2.293A1 1 0 007 2H1.5zM1 3.5a.5.5 0 01.5-.5H7l1.707 1.707A1 1 0 009.414 5H14.5a.5.5 0 01.5.5v7a.5.5 0 01-.5.5h-13a.5.5 0 01-.5-.5v-9z" />
    </svg>
)

const FolderClosedIcon = () => (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
        <path d="M1.5 2A1.5 1.5 0 000 3.5v9A1.5 1.5 0 001.5 14h13a1.5 1.5 0 001.5-1.5v-7A1.5 1.5 0 0014.5 4H9.414L7.707 2.293A1 1 0 007 2H1.5z" />
    </svg>
)

const FileIcon = () => (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
        <path d="M4 0a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V5.414A1.5 1.5 0 0013.414 4L10 .586A1.5 1.5 0 009.414 0H4zm0 1h5v3.5A1.5 1.5 0 0010.5 6H13v8a1 1 0 01-1 1H4a1 1 0 01-1-1V2a1 1 0 011-1z" />
    </svg>
)

export default function Sidebar({ activeMenu = 'Menus', onMenuClick }: SidebarProps) {
    const menuItems = [
        { id: 'systems', type: 'folder', label: 'Systems', hasChildren: true },
        { id: 'systemCode', type: 'file', label: 'System Code', indent: true },
        { id: 'properties', type: 'file', label: 'Properties', indent: true },
        { id: 'menus', type: 'file', label: 'Menus', indent: true },
        { id: 'apiList', type: 'file', label: 'API List', indent: true },
        { id: 'usersGroup', type: 'folder', label: 'Users & Group', hasChildren: false },
        { id: 'competition', type: 'folder', label: 'Competition', hasChildren: false },
    ]

    const renderIcon = (item: typeof menuItems[0], isActive: boolean) => {
        if (item.type === 'folder') {
            return isActive ? <FolderOpenIcon /> : <FolderClosedIcon />
        }
        return <FileIcon />
    }

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
                                } ${item.indent ? 'pl-10' : ''}`}
                        >
                            <span className={`flex-shrink-0 ${isActive ? '' : 'opacity-80'}`}>
                                {renderIcon(item, isActive)}
                            </span>
                            <span className="text-sm">{item.label}</span>
                        </button>
                    )
                })}
            </nav>
        </div>
    )
}
