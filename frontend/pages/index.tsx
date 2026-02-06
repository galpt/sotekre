import React, { useState } from 'react'
import Sidebar from '../components/Sidebar'
import MenuTree from '../components/MenuTree'
import DetailsPanel from '../components/DetailsPanel'

type MenuItem = {
    id: string
    name: string
    depth: number
    parentData?: string
    children?: MenuItem[]
}

// Sample data matching the Figma design
const sampleMenuData: MenuItem[] = [
    {
        id: '1',
        name: 'system management',
        depth: 1,
        children: [
            {
                id: '2',
                name: 'System Management',
                depth: 2,
                parentData: 'system management',
                children: [
                    {
                        id: '3',
                        name: 'Systems',
                        depth: 3,
                        parentData: 'System Management',
                        children: [
                            {
                                id: '56320ee9-6af6-11ed-a7ba-f220afe5e4a9',
                                name: 'System Code',
                                depth: 3,
                                parentData: 'Systems',
                                children: [
                                    {
                                        id: '5',
                                        name: 'Code Registration',
                                        depth: 4,
                                        parentData: 'System Code',
                                    },
                                    {
                                        id: '6',
                                        name: 'Code Registration - 2',
                                        depth: 4,
                                        parentData: 'System Code',
                                    },
                                ],
                            },
                            {
                                id: '7',
                                name: 'Properties',
                                depth: 4,
                                parentData: 'Systems',
                            },
                        ],
                    },
                    {
                        id: '8',
                        name: 'Menus',
                        depth: 3,
                        parentData: 'System Management',
                        children: [
                            {
                                id: '9',
                                name: 'Menu Registration',
                                depth: 4,
                                parentData: 'Menus',
                            },
                        ],
                    },
                    {
                        id: '10',
                        name: 'API List',
                        depth: 3,
                        parentData: 'System Management',
                        children: [
                            {
                                id: '11',
                                name: 'API Registration',
                                depth: 4,
                                parentData: 'API List',
                            },
                            {
                                id: '12',
                                name: 'API Edit',
                                depth: 4,
                                parentData: 'API List',
                            },
                        ],
                    },
                ],
            },
            {
                id: '13',
                name: 'Users & Groups',
                depth: 2,
                parentData: 'system management',
                children: [
                    {
                        id: '14',
                        name: 'Users',
                        depth: 3,
                        parentData: 'Users & Groups',
                        children: [
                            {
                                id: '15',
                                name: 'User Account Registration',
                                depth: 4,
                                parentData: 'Users',
                            },
                        ],
                    },
                    {
                        id: '16',
                        name: 'Groups',
                        depth: 3,
                        parentData: 'Users & Groups',
                        children: [
                            {
                                id: '17',
                                name: 'User Group Registration',
                                depth: 4,
                                parentData: 'Groups',
                            },
                        ],
                    },
                ],
            },
            {
                id: '18',
                name: '사용자 승인',
                depth: 2,
                parentData: 'system management',
                children: [
                    {
                        id: '19',
                        name: '사용자 승인 상세',
                        depth: 3,
                        parentData: '사용자 승인',
                    },
                ],
            },
        ],
    },
]

export default function Home() {
    const [selectedItem, setSelectedItem] = useState<MenuItem | null>(null)
    const [expandedAll, setExpandedAll] = useState<boolean | undefined>(undefined)
    const [selectedCategory, setSelectedCategory] = useState('system management')

    const handleExpandAll = () => {
        setExpandedAll(true)
    }

    const handleCollapseAll = () => {
        setExpandedAll(false)
    }

    const handleSave = (item: MenuItem) => {
        console.log('Saving item:', item)
        // Add your save logic here
    }

    return (
        <div className="flex h-screen bg-gray-50">
            {/* Sidebar */}
            <Sidebar activeMenu="Menus" />

            {/* Main Content */}
            <div className="ml-[188px] flex-1 flex">
                {/* Left Panel - Menu Tree */}
                <div className="flex-1 flex flex-col">
                    {/* Header */}
                    <div className="bg-white border-b px-6 py-4">
                        <div className="flex items-center gap-2 mb-4">
                            <span className="text-gray-400 text-sm">📁 Menus</span>
                        </div>
                        <div className="flex items-center gap-3 mb-4">
                            <div className="w-10 h-10 bg-[#0D47A1] rounded-lg flex items-center justify-center">
                                <svg
                                    width="24"
                                    height="24"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="white"
                                    strokeWidth="2"
                                >
                                    <rect x="3" y="3" width="7" height="7" />
                                    <rect x="14" y="3" width="7" height="7" />
                                    <rect x="14" y="14" width="7" height="7" />
                                    <rect x="3" y="14" width="7" height="7" />
                                </svg>
                            </div>
                            <h1 className="text-2xl font-semibold text-gray-900">Menus</h1>
                        </div>

                        {/* Dropdown and Buttons */}
                        <div className="flex items-center gap-4">
                            <div className="flex-1">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Menu</label>
                                <select
                                    value={selectedCategory}
                                    onChange={(e) => setSelectedCategory(e.target.value)}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="system management">system management</option>
                                </select>
                            </div>
                            <div className="flex gap-2 mt-6">
                                <button
                                    onClick={handleExpandAll}
                                    className="px-4 py-2 bg-gray-800 hover:bg-gray-900 text-white text-sm font-medium rounded-lg transition-colors"
                                >
                                    Expand All
                                </button>
                                <button
                                    onClick={handleCollapseAll}
                                    className="px-4 py-2 border border-gray-300 hover:bg-gray-50 text-gray-700 text-sm font-medium rounded-lg transition-colors"
                                >
                                    Collapse All
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Tree View */}
                    <div className="flex-1 overflow-auto bg-white p-4">
                        <MenuTree
                            items={sampleMenuData}
                            selectedId={selectedItem?.id}
                            onSelectItem={setSelectedItem}
                            expandedAll={expandedAll}
                        />
                    </div>
                </div>

                {/* Right Panel - Details */}
                <div className="w-[400px] bg-white border-l">
                    <DetailsPanel selectedItem={selectedItem} onSave={handleSave} />
                </div>
            </div>
        </div>
    )
}
