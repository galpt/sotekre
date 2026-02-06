import React, { useState, useEffect } from 'react'
import Sidebar from '../components/Sidebar'
import MenuTree from '../components/MenuTree'
import DetailsPanel from '../components/DetailsPanel'
import { menuService, MenuNode as APIMenuNode } from '../services/menuService'

type MenuItem = {
    id: string
    name: string
    depth: number
    parentData?: string
    children?: MenuItem[]
}

// Helper function to calculate depth and convert API data to UI format
const convertAPIToUI = (node: APIMenuNode, depth: number = 1, parentTitle?: string): MenuItem => {
    return {
        id: node.id.toString(),
        name: node.title,
        depth,
        parentData: parentTitle,
        children: node.children?.map(child => convertAPIToUI(child, depth + 1, node.title)),
    }
}

// Helper to find parent title given parent_id
const findParentTitle = (nodes: APIMenuNode[], parentId: number): string | undefined => {
    for (const node of nodes) {
        if (node.id === parentId) return node.title
        if (node.children) {
            const found = findParentTitle(node.children, parentId)
            if (found) return found
        }
    }
    return undefined
}

export default function Home() {
    const [menuData, setMenuData] = useState<MenuItem[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [selectedItem, setSelectedItem] = useState<MenuItem | null>(null)
    const [expandedAll, setExpandedAll] = useState<boolean | undefined>(undefined)
    const [selectedCategory, setSelectedCategory] = useState('system management')

    // Load menus from API
    useEffect(() => {
        loadMenus()
    }, [])

    const loadMenus = async () => {
        try {
            setLoading(true)
            setError(null)
            const data = await menuService.getMenus()
            const converted = data.map(node => convertAPIToUI(node))
            setMenuData(converted)
        } catch (err: any) {
            console.error('Failed to load menus:', err)
            setError(err.message || 'Failed to load menus')
        } finally {
            setLoading(false)
        }
    }

    const handleExpandAll = () => {
        setExpandedAll(true)
    }

    const handleCollapseAll = () => {
        setExpandedAll(false)
    }

    const handleSave = async (item: MenuItem) => {
        try {
            // Here you can implement update logic
            console.log('Saving item:', item)
            await menuService.updateMenu(parseInt(item.id), {
                title: item.name,
            })
            await loadMenus() // Reload after save
        } catch (err: any) {
            console.error('Failed to save:', err)
            alert('Failed to save: ' + err.message)
        }
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
                        {loading ? (
                            <div className="flex items-center justify-center h-full">
                                <div className="text-gray-500">Loading menus...</div>
                            </div>
                        ) : error ? (
                            <div className="flex items-center justify-center h-full">
                                <div className="text-red-500">Error: {error}</div>
                            </div>
                        ) : menuData.length === 0 ? (
                            <div className="flex items-center justify-center h-full">
                                <div className="text-gray-400">No menus found</div>
                            </div>
                        ) : (
                            <MenuTree
                                items={menuData}
                                selectedId={selectedItem?.id}
                                onSelectItem={setSelectedItem}
                                expandedAll={expandedAll}
                            />
                        )}
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
