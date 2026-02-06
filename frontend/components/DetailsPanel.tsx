import React, { useState, useEffect } from 'react'

type MenuItem = {
    id: string
    name: string
    depth: number
    parentData?: string
    url?: string
}

type DetailsPanelProps = {
    selectedItem: MenuItem | null
    onSave?: (item: MenuItem) => void
    onDelete?: (id: string) => void
}

export default function DetailsPanel({ selectedItem, onSave, onDelete }: DetailsPanelProps) {
    const [editedName, setEditedName] = useState('')
    const [editedUrl, setEditedUrl] = useState('')
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

    useEffect(() => {
        if (selectedItem) {
            setEditedName(selectedItem.name)
            setEditedUrl(selectedItem.url || '')
        }
    }, [selectedItem])

    if (!selectedItem) {
        return (
            <div className="flex items-center justify-center h-full text-gray-400">
                <p>Select a menu item to view details</p>
            </div>
        )
    }

    const handleSave = () => {
        if (editedName.trim()) {
            onSave?.({
                ...selectedItem,
                name: editedName,
                url: editedUrl.trim() || undefined // Convert empty string to undefined
            })
        }
    }

    const handleDelete = () => {
        onDelete?.(selectedItem.id)
        setShowDeleteConfirm(false)
    }

    return (
        <div className="p-6">
            <div className="space-y-6">
                {/* Menu ID */}
                <div>
                    <label className="block text-sm font-medium text-gray-600 mb-2">
                        Menu ID
                    </label>
                    <div className="bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 text-sm text-gray-700">
                        {selectedItem.id}
                    </div>
                </div>

                {/* Depth */}
                <div>
                    <label className="block text-sm font-medium text-gray-600 mb-2">
                        Depth
                    </label>
                    <div className="bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 text-sm text-gray-700">
                        {selectedItem.depth}
                    </div>
                </div>

                {/* Parent Data */}
                <div>
                    <label className="block text-sm font-medium text-gray-600 mb-2">
                        Parent Data
                    </label>
                    <div className="bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 text-sm text-gray-700">
                        {selectedItem.parentData || '-'}
                    </div>
                </div>

                {/* Name - Editable */}
                <div>
                    <label className="block text-sm font-medium text-gray-600 mb-2">
                        Name
                    </label>
                    <input
                        type="text"
                        value={editedName}
                        onChange={(e) => setEditedName(e.target.value)}
                        className="w-full border border-gray-300 rounded-lg px-4 py-2.5 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>

                {/* URL - Editable */}
                <div>
                    <label className="block text-sm font-medium text-gray-600 mb-2">
                        URL
                    </label>
                    <input
                        type="text"
                        value={editedUrl}
                        onChange={(e) => setEditedUrl(e.target.value)}
                        placeholder="/path/to/page"
                        className="w-full border border-gray-300 rounded-lg px-4 py-2.5 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>

                {/* Action Buttons */}
                <div className="pt-4 space-y-3">
                    <button
                        onClick={handleSave}
                        className="w-full bg-[#0D47A1] hover:bg-[#083A89] text-white font-medium py-3 px-6 rounded-lg transition-colors shadow-sm"
                    >
                        Save
                    </button>
                    <button
                        onClick={() => setShowDeleteConfirm(true)}
                        className="w-full bg-red-600 hover:bg-red-700 text-white font-medium py-3 px-6 rounded-lg transition-colors shadow-sm"
                    >
                        Delete
                    </button>
                </div>
            </div>

            {/* Delete Confirmation Modal */}
            {showDeleteConfirm && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={() => setShowDeleteConfirm(false)}>
                    <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4" onClick={(e) => e.stopPropagation()}>
                        <h3 className="text-lg font-semibold text-gray-900 mb-2">Delete Menu Item</h3>
                        <p className="text-gray-600 mb-6">
                            Are you sure you want to delete "{selectedItem.name}"? This will also delete all child menu items.
                        </p>
                        <div className="flex gap-3">
                            <button
                                onClick={() => setShowDeleteConfirm(false)}
                                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleDelete}
                                className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                            >
                                Delete
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
