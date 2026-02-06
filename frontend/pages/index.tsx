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
    const [sidebarOpen, setSidebarOpen] = useState(false)
    const [detailsPanelOpen, setDetailsPanelOpen] = useState(false)

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
            console.log('Saving item:', item)
            // Convert empty string url to undefined to properly handle NULL in database
            const updateData: { title: string; url?: string } = {
                title: item.name,
            }
            if (item.url && item.url.trim()) {
                updateData.url = item.url.trim()
            }

            await menuService.updateMenu(parseInt(item.id), updateData)
            await loadMenus() // Reload after save
        } catch (err: any) {
            console.error('Failed to save:', err)
            alert('Failed to save: ' + err.message)
        }
    }

    const handleDelete = async (id: string) => {
        try {
            console.log(`[DELETE] Starting delete for ID: ${id}`)
            const numericId = parseInt(id)
            if (isNaN(numericId)) {
                throw new Error(`Invalid ID: ${id}`)
            }

            console.log(`[DELETE] Calling API to delete ID: ${numericId}`)
            await menuService.deleteMenu(numericId)
            console.log(`[DELETE] API call successful, clearing selection`)

            setSelectedItem(null)

            console.log(`[DELETE] Reloading menus...`)
            await loadMenus()
            console.log(`[DELETE] Reload complete`)
        } catch (err: any) {
            console.error('[DELETE] Failed to delete:', err)
            alert('Failed to delete: ' + (err.response?.data?.error || err.message))
        }
    }

    const handleMoveItem = async (itemId: string, newParentId: string | null, newOrder: number) => {
        try {
            console.log(`[MOVE] Moving item ${itemId} to parent ${newParentId} at order ${newOrder}`)
            const numericId = parseInt(itemId)
            const numericParentId = newParentId ? parseInt(newParentId) : null

            await menuService.moveMenu(numericId, numericParentId, newOrder)
            console.log(`[MOVE] Move successful, reloading...`)
            await loadMenus()
            console.log(`[MOVE] Reload complete`)
        } catch (err: any) {
            console.error('[MOVE] Failed to move item:', err)
            alert('Failed to move item: ' + (err.response?.data?.error || err.message))
            // Reload to revert UI to actual state
            await loadMenus()
        }
    }

    const handleCreateMenu = async () => {
        if (!newMenuTitle.trim()) {
            alert('Menu title is required')
            return
        }

        try {
            const createData: {
                title: string
                url?: string
                parent_id?: number
            } = {
                title: newMenuTitle.trim(),
            }

            // Only add url if it's not empty
            if (newMenuUrl && newMenuUrl.trim()) {
                createData.url = newMenuUrl.trim()
            }

            // Only add parent_id if specified
            if (newMenuParentId) {
                createData.parent_id = parseInt(newMenuParentId)
            }

            await menuService.createMenu(createData)
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

    const handleSelectItem = (item: MenuItem) => {
        setSelectedItem(item)
        setDetailsPanelOpen(true)
    }

    return (
        <div className="flex h-screen bg-gray-50">
            {/* Mobile Menu Button */}
            <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="md:hidden fixed top-4 left-4 z-50 p-2 bg-[#0D47A1] text-white rounded-lg"
            >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
            </button>

            {/* Sidebar */}
            <div className={`fixed md:static inset-0 z-40 transform ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'
                } md:translate-x-0 transition-transform duration-300`}>
                <div className="md:hidden fixed inset-0 bg-black bg-opacity-50" onClick={() => setSidebarOpen(false)} />
                <div className="relative">
                    <Sidebar activeMenu="Menus" />
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 flex flex-col md:flex-row md:ml-[188px]">
                {/* Left Panel - Menu Tree */}
                <div className="flex-1 flex flex-col w-full md:w-auto">
                    {/* Header */}
                    <div className="bg-white border-b px-4 md:px-6 py-4">
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
                            <div className="flex flex-wrap gap-2">
                                <button
                                    onClick={() => { setNewMenuParentId(''); setShowCreateModal(true); }}
                                    className="px-3 md:px-4 py-2 bg-[#0D47A1] hover:bg-[#083A89] text-white text-xs md:text-sm font-medium rounded-lg transition-colors"
                                >
                                    + Create
                                </button>
                                <button
                                    onClick={handleExpandAll}
                                    className="px-3 md:px-4 py-2 bg-gray-800 hover:bg-gray-900 text-white text-xs md:text-sm font-medium rounded-lg transition-colors"
                                >
                                    Expand
                                </button>
                                <button
                                    onClick={handleCollapseAll}
                                    className="px-3 md:px-4 py-2 border border-gray-300 hover:bg-gray-50 text-gray-700 text-xs md:text-sm font-medium rounded-lg transition-colors"
                                >
                                    Collapse
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Search Bar */}
                    <div className="bg-white border-b px-4 md:px-6 py-3">
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
                        <div className="flex flex-col items-center justify-center h-full space-y-4">
                            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#0D47A1]"></div>
                            <div className="text-gray-500 text-sm">Loading menus...</div>
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
                            onSelectItem={handleSelectItem}
                            onAddChild={handleAddChild}
                            expandedAll={expandedAll}
                            onMoveItem={handleMoveItem}
                        />
                    )}
                    </div>
                </div>

                {/* Right Panel - Details (Desktop) */}
                <div className="hidden md:block md:w-[400px] bg-white border-l">
                    <DetailsPanel selectedItem={selectedItem} onSave={handleSave} onDelete={handleDelete} />
                </div>

                {/* Right Panel - Details (Mobile Slide-over) */}
                {detailsPanelOpen && (
                    <div className="md:hidden fixed inset-0 z-50">
                        <div className="absolute inset-0 bg-black bg-opacity-50" onClick={() => setDetailsPanelOpen(false)} />
                        <div className="absolute right-0 top-0 bottom-0 w-full max-w-sm bg-white shadow-xl">
                            <div className="flex items-center justify-between p-4 border-b">
                                <h3 className="text-lg font-semibold">Menu Details</h3>
                                <button
                                    onClick={() => setDetailsPanelOpen(false)}
                                    className="p-2 hover:bg-gray-100 rounded-lg"
                                >
                                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            </div>
                            <div className="overflow-auto h-[calc(100%-65px)]">
                                <DetailsPanel
                                    selectedItem={selectedItem}
                                    onSave={async (id, data) => {
                                        await handleSave(id, data)
                                        setDetailsPanelOpen(false)
                                    }}
                                    onDelete={async (id) => {
                                        await handleDelete(id)
                                        setDetailsPanelOpen(false)
                                    }}
                                />
                            </div>
                        </div>
                    </div>
                )}
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
