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
    url?: string
    children?: MenuItem[]
}

// Helper function to calculate depth and convert API data to UI format
const convertAPIToUI = (node: APIMenuNode, depth: number = 1, parentTitle?: string): MenuItem => {
    return {
        id: node.id.toString(),
        name: node.title,
        url: node.url,
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
    const [showCreateModal, setShowCreateModal] = useState(false)
    const [newMenuTitle, setNewMenuTitle] = useState('')
    const [newMenuUrl, setNewMenuUrl] = useState('')
    const [newMenuParentId, setNewMenuParentId] = useState<string>('')
    const [searchQuery, setSearchQuery] = useState('')

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
                url: item.url,
            })
            await loadMenus() // Reload after save
        } catch (err: any) {
            console.error('Failed to save:', err)
            alert('Failed to save: ' + err.message)
        }
    }

    const handleDelete = async (id: string) => {
        try {
            await menuService.deleteMenu(parseInt(id))
            setSelectedItem(null)
            await loadMenus() // Reload after delete
        } catch (err: any) {
            console.error('Failed to delete:', err)
            alert('Failed to delete: ' + err.message)
        }
    }

    const handleCreateMenu = async () => {
        if (!newMenuTitle.trim()) {
            alert('Menu title is required')
            return
        }

        try {
            await menuService.createMenu({
                title: newMenuTitle,
                url: newMenuUrl || undefined,
                parent_id: newMenuParentId ? parseInt(newMenuParentId) : undefined,
            })
            setShowCreateModal(false)
            setNewMenuTitle('')
            setNewMenuUrl('')
            setNewMenuParentId('')
            await loadMenus() // Reload after create
        } catch (err: any) {
            console.error('Failed to create menu:', err)
            alert('Failed to create menu: ' + err.message)
        }
    }

    const handleAddChild = (parentItem: MenuItem) => {
        setNewMenuParentId(parentItem.id)
        setNewMenuTitle('')
        setNewMenuUrl('')
        setShowCreateModal(true)
    }

    // Get all menu items as flat list for parent selection
    const getAllMenuItems = (items: MenuItem[]): MenuItem[] => {
        let result: MenuItem[] = []
        items.forEach(item => {
            result.push(item)
            if (item.children) {
                result = result.concat(getAllMenuItems(item.children))
            }
        })
        return result
    }

    // Filter menu items based on search query
    const filterMenuItems = (items: MenuItem[], query: string): MenuItem[] => {
        if (!query.trim()) return items

        const lowerQuery = query.toLowerCase()
        return items.filter(item => {
            const matchesName = item.name.toLowerCase().includes(lowerQuery)
            const matchesUrl = item.url?.toLowerCase().includes(lowerQuery)
            const hasMatchingChildren = item.children && filterMenuItems(item.children, query).length > 0

            if (matchesName || matchesUrl || hasMatchingChildren) {
                return {
                    ...item,
                    children: item.children ? filterMenuItems(item.children, query) : undefined
                }
            }
            return false
        }).map(item => ({
            ...item,
            children: item.children ? filterMenuItems(item.children, query) : undefined
        }))
    }

    const filteredMenuData = filterMenuItems(menuData, searchQuery)

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
                                    onClick={() => { setNewMenuParentId(''); setShowCreateModal(true); }}
                                    className="px-4 py-2 bg-[#0D47A1] hover:bg-[#083A89] text-white text-sm font-medium rounded-lg transition-colors"
                                >
                                    + Create Menu
                                </button>
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

                    {/* Search Bar */}
                    <div className="bg-white border-b px-6 py-3">
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            placeholder="Search menus..."
                            className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>

                    {/* Tree View */}
                    <div className="flex-1 overflow-auto bg-white p-4">{loading ? (
                        <div className="flex items-center justify-center h-full">
                            <div className="text-gray-500">Loading menus...</div>
                        </div>
                    ) : error ? (
                        <div className="flex items-center justify-center h-full">
                            <div className="text-red-500">Error: {error}</div>
                        </div>
                    ) : filteredMenuData.length === 0 ? (
                        <div className="flex items-center justify-center h-full">
                            <div className="text-gray-400">
                                {searchQuery ? 'No menus match your search' : 'No menus found'}
                            </div>
                        </div>
                    ) : (
                        <MenuTree
                            items={filteredMenuData}
                            selectedId={selectedItem?.id}
                            onSelectItem={setSelectedItem}
                            onAddChild={handleAddChild}
                            expandedAll={expandedAll}
                        />
                    )}
                    </div>
                </div>

                {/* Right Panel - Details */}
                <div className="w-[400px] bg-white border-l">
                    <DetailsPanel selectedItem={selectedItem} onSave={handleSave} onDelete={handleDelete} />
                </div>
            </div>
            {/* Create Menu Modal */}
            {showCreateModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={() => setShowCreateModal(false)}>
                    <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4" onClick={(e) => e.stopPropagation()}>
                        <h3 className="text-lg font-semibold text-gray-900 mb-4">Create New Menu</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">Title *</label>
                                <input
                                    type="text"
                                    value={newMenuTitle}
                                    onChange={(e) => setNewMenuTitle(e.target.value)}
                                    placeholder="Menu title"
                                    className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">URL</label>
                                <input
                                    type="text"
                                    value={newMenuUrl}
                                    onChange={(e) => setNewMenuUrl(e.target.value)}
                                    placeholder="/path/to/page"
                                    className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">Parent Menu (optional)</label>
                                <select
                                    value={newMenuParentId}
                                    onChange={(e) => setNewMenuParentId(e.target.value)}
                                    className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="">-- Root Level --</option>
                                    {getAllMenuItems(menuData).map(item => (
                                        <option key={item.id} value={item.id}>
                                            {"  ".repeat(item.depth - 1)}{item.name}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                        <div className="flex gap-3 mt-6">
                            <button
                                onClick={() => setShowCreateModal(false)}
                                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleCreateMenu}
                                className="flex-1 px-4 py-2 bg-[#0D47A1] hover:bg-[#083A89] text-white rounded-lg transition-colors"
                            >
                                Create
                            </button>
                        </div>
                    </div>
                </div>
            )}        </div>
    )
}
