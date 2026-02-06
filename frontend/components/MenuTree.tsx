import React, { useState } from 'react'

type MenuItem = {
    id: string
    name: string
    depth: number
    parentData?: string
    children?: MenuItem[]
}

type MenuTreeProps = {
    items: MenuItem[]
    selectedId?: string | null
    onSelectItem?: (item: MenuItem) => void
    onAddChild?: (parentItem: MenuItem) => void
    expandedAll?: boolean
}

function TreeNode({
    item,
    selectedId,
    onSelectItem,
    onAddChild,
    depth = 0,
    forceExpanded
}: {
    item: MenuItem
    selectedId?: string | null
    onSelectItem?: (item: MenuItem) => void
    onAddChild?: (parentItem: MenuItem) => void
    depth?: number
    forceExpanded?: boolean
}) {
    const [isExpanded, setIsExpanded] = useState(false)
    const hasChildren = item.children && item.children.length > 0
    const isSelected = item.id === selectedId
    const expanded = forceExpanded !== undefined ? forceExpanded : isExpanded

    return (
        <div>
            <div
                className={`flex items-center gap-2 py-1 px-2 cursor-pointer hover:bg-gray-50 ${isSelected ? 'bg-blue-50' : ''
                    }`}
                style={{ paddingLeft: `${depth * 24 + 8}px` }}
                onClick={() => onSelectItem?.(item)}
            >
                {hasChildren ? (
                    <button
                        onClick={(e) => {
                            e.stopPropagation()
                            setIsExpanded(!isExpanded)
                        }}
                        className="w-4 h-4 flex items-center justify-center text-gray-500 hover:text-gray-700"
                    >
                        <svg
                            width="12"
                            height="12"
                            viewBox="0 0 12 12"
                            fill="none"
                            className={`transform transition-transform ${expanded ? 'rotate-90' : ''}`}
                        >
                            <path
                                d="M4 2L8 6L4 10"
                                stroke="currentColor"
                                strokeWidth="1.5"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                            />
                        </svg>
                    </button>
                ) : (
                    <div className="w-4 h-4"></div>
                )}
                <span className="text-sm text-gray-800 flex-1">{item.name}</span>
                {hasChildren && expanded && (
                    <button
                        onClick={(e) => {
                            e.stopPropagation()
                            onAddChild?.(item)
                        }}
                        className="w-5 h-5 bg-[#0D47A1] hover:bg-[#083A89] text-white rounded flex items-center justify-center transition-colors"
                        title="Add child menu"
                    >
                        <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                            <path d="M6 2V10M2 6H10" stroke="white" strokeWidth="1.5" strokeLinecap="round" />
                        </svg>
                    </button>
                )}
                {hasChildren && (
                    <span className="ml-1 w-5 h-5 bg-[#0D47A1] text-white rounded-full flex items-center justify-center text-xs font-medium">
                        {item.children!.length}
                    </span>
                )}
            </div>
            {hasChildren && expanded && (
                <div>
                    {item.children!.map((child) => (
                        <TreeNode
                            key={child.id}
                            item={child}
                            selectedId={selectedId}
                            onSelectItem={onSelectItem}
                            onAddChild={onAddChild}
                            depth={depth + 1}
                            forceExpanded={forceExpanded}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}

export default function MenuTree({ items, selectedId, onSelectItem, onAddChild, expandedAll }: MenuTreeProps) {
    return (
        <div className="w-full">
            {items.map((item) => (
                <TreeNode
                    key={item.id}
                    item={item}
                    selectedId={selectedId}
                    onSelectItem={onSelectItem}
                    onAddChild={onAddChild}
                    depth={0}
                    forceExpanded={expandedAll}
                />
            ))}
        </div>
    )
}
