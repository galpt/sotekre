import React, { useState } from 'react'
import {
    DndContext,
    closestCenter,
    KeyboardSensor,
    PointerSensor,
    useSensor,
    useSensors,
    DragEndEvent,
    DragStartEvent,
    DragOverlay,
} from '@dnd-kit/core'
import {
    SortableContext,
    sortableKeyboardCoordinates,
    verticalListSortingStrategy,
    useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

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
    onMoveItem?: (itemId: string, newParentId: string | null, newOrder: number) => Promise<void>
}

function TreeNode({
    item,
    selectedId,
    onSelectItem,
    onAddChild,
    depth = 0,
    forceExpanded,
    isDragging = false,
}: {
    item: MenuItem
    selectedId?: string | null
    onSelectItem?: (item: MenuItem) => void
    onAddChild?: (parentItem: MenuItem) => void
    depth?: number
    forceExpanded?: boolean
    isDragging?: boolean
}) {
    const [isExpanded, setIsExpanded] = useState(false)
    const hasChildren = item.children && item.children.length > 0
    const isSelected = item.id === selectedId
    const expanded = forceExpanded !== undefined ? forceExpanded : isExpanded

    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
        isDragging: isSortableDragging,
    } = useSortable({ id: item.id })

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isSortableDragging ? 0.5 : 1,
    }

    return (
        <div ref={setNodeRef} style={style}>
            <div
                className={`flex items-center gap-2 py-1.5 px-2 cursor-pointer hover:bg-gray-100 rounded-md transition-all duration-150 ${isSelected ? 'bg-blue-50 ring-1 ring-blue-200' : ''
                    } ${isSortableDragging ? 'shadow-lg z-50 opacity-50' : ''}`}
                style={{ paddingLeft: `${depth * 24 + 8}px` }}
                onClick={() => onSelectItem?.(item)}
            >
                {/* Drag Handle */}
                <button
                    {...attributes}
                    {...listeners}
                    className="w-5 h-5 flex items-center justify-center text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded cursor-grab active:cursor-grabbing transition-colors"
                    onClick={(e) => e.stopPropagation()}
                    title="Drag to reorder"
                >
                    <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor">
                        <circle cx="3" cy="3" r="1.2" />
                        <circle cx="9" cy="3" r="1.2" />
                        <circle cx="3" cy="6" r="1.2" />
                        <circle cx="9" cy="6" r="1.2" />
                        <circle cx="3" cy="9" r="1.2" />
                        <circle cx="9" cy="9" r="1.2" />
                    </svg>
                </button>

                {/* Expand/Collapse */}
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

                {/* Item Name */}
                <span className="text-sm text-gray-800 flex-1">{item.name}</span>

                {/* Add Child Button */}
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

                {/* Children Count Badge */}
                {hasChildren && (
                    <span className="ml-1 w-5 h-5 bg-[#0D47A1] text-white rounded-full flex items-center justify-center text-xs font-medium">
                        {item.children!.length}
                    </span>
                )}
            </div>

            {/* Child Items */}
            {hasChildren && expanded && (
                <div>
                    <SortableContext items={item.children!.map(c => c.id)} strategy={verticalListSortingStrategy}>
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
                    </SortableContext>
                </div>
            )}
        </div>
    )
}

export default function MenuTree({ items, selectedId, onSelectItem, onAddChild, expandedAll, onMoveItem }: MenuTreeProps) {
    const [activeId, setActiveId] = useState<string | null>(null)
    const [draggedItem, setDraggedItem] = useState<MenuItem | null>(null)

    const sensors = useSensors(
        useSensor(PointerSensor, {
            activationConstraint: {
                distance: 8,
            },
        }),
        useSensor(KeyboardSensor, {
            coordinateGetter: sortableKeyboardCoordinates,
        })
    )

    // Flatten tree to find items and their parents
    const flattenTree = (items: MenuItem[], parent: MenuItem | null = null): Array<{ item: MenuItem; parent: MenuItem | null }> => {
        const result: Array<{ item: MenuItem; parent: MenuItem | null }> = []
        for (const item of items) {
            result.push({ item, parent })
            if (item.children) {
                result.push(...flattenTree(item.children, item))
            }
        }
        return result
    }

    const findItemById = (id: string | null): MenuItem | null => {
        if (!id) return null
        const flatItems = flattenTree(items)
        const found = flatItems.find(({ item }) => item.id === id)
        return found?.item || null
    }

    const findParentById = (childId: string): MenuItem | null => {
        const flatItems = flattenTree(items)
        const found = flatItems.find(({ item }) => item.id === childId)
        return found?.parent || null
    }

    const findItemIndex = (parentItems: MenuItem[], itemId: string): number => {
        return parentItems.findIndex(item => item.id === itemId)
    }

    const handleDragStart = (event: DragStartEvent) => {
        const { active } = event
        setActiveId(active.id as string)
        const item = findItemById(active.id as string)
        setDraggedItem(item)
    }

    const handleDragEnd = async (event: DragEndEvent) => {
        const { active, over } = event
        setActiveId(null)
        setDraggedItem(null)

        if (!over || active.id === over.id) {
            return
        }

        const activeItem = findItemById(active.id as string)
        const overItem = findItemById(over.id as string)
        const activeParent = findParentById(active.id as string)
        const overParent = findParentById(over.id as string)

        if (!activeItem || !overItem) {
            return
        }

        // Get the list where both items belong
        const activeParentItems = activeParent?.children || items
        const overParentItems = overParent?.children || items

        // Check if reordering within same parent
        const sameParent = activeParent?.id === overParent?.id || (!activeParent && !overParent)

        if (sameParent) {
            // Reorder within same parent
            const oldIndex = findItemIndex(activeParentItems, active.id as string)
            const newIndex = findItemIndex(overParentItems, over.id as string)

            if (oldIndex !== newIndex && onMoveItem) {
                try {
                    await onMoveItem(
                        activeItem.id,
                        activeParent?.id || null,
                        newIndex
                    )
                } catch (error) {
                    console.error('Failed to reorder item:', error)
                }
            }
        } else {
            // Move to different parent
            const newIndex = findItemIndex(overParentItems, over.id as string)

            if (onMoveItem) {
                try {
                    await onMoveItem(
                        activeItem.id,
                        overParent?.id || null,
                        newIndex
                    )
                } catch (error) {
                    console.error('Failed to  move item:', error)
                }
            }
        }
    }

    return (
        <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
        >
            <div className="w-full">
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
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
                </SortableContext>
            </div>
            <DragOverlay>
                {draggedItem ? (
                    <div className="bg-blue-50 shadow-2xl rounded-md p-2 border-2 border-[#0D47A1] ring-2 ring-blue-200 opacity-95">
                        <div className="flex items-center gap-2">
                            <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor" className="text-gray-600">
                                <circle cx="3" cy="3" r="1.2" />
                                <circle cx="9" cy="3" r="1.2" />
                                <circle cx="3" cy="6" r="1.2" />
                                <circle cx="9" cy="6" r="1.2" />
                                <circle cx="3" cy="9" r="1.2" />
                                <circle cx="9" cy="9" r="1.2" />
                            </svg>
                            <span className="text-sm font-medium text-gray-800">{draggedItem.name}</span>
                        </div>
                    </div>
                ) : null}
            </DragOverlay>
        </DndContext>
    )
}
