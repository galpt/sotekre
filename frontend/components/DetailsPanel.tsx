import React from 'react'

type MenuItem = {
    id: string
    name: string
    depth: number
    parentData?: string
}

type DetailsPanelProps = {
    selectedItem: MenuItem | null
    onSave?: (item: MenuItem) => void
}

export default function DetailsPanel({ selectedItem, onSave }: DetailsPanelProps) {
    if (!selectedItem) {
        return (
            <div className="flex items-center justify-center h-full text-gray-400">
                <p>Select a menu item to view details</p>
            </div>
        )
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

                {/* Name */}
                <div>
                    <label className="block text-sm font-medium text-gray-600 mb-2">
                        Name
                    </label>
                    <div className="bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 text-sm text-gray-700">
                        {selectedItem.name}
                    </div>
                </div>

                {/* Save Button */}
                <div className="pt-4">
                    <button
                        onClick={() => onSave?.(selectedItem)}
                        className="w-full bg-[#0D47A1] hover:bg-[#083A89] text-white font-medium py-3 px-6 rounded-lg transition-colors shadow-sm"
                    >
                        Save
                    </button>
                </div>
            </div>
        </div>
    )
}
