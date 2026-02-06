import React, { useEffect, useState } from 'react'

type MenuNode = {
    id: number
    title: string
    url?: string | null
    parent_id?: number | null
    order?: number
    children?: MenuNode[]
}

const fetchMenus = async (): Promise<MenuNode[]> => {
    const res = await fetch('/api/menus')
    if (!res.ok) throw new Error('Failed to fetch')
    const body = await res.json()
    return body.data || []
}

export default function Home() {
    const [tree, setTree] = useState<MenuNode[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [search, setSearch] = useState('')
    const [titleInput, setTitleInput] = useState('')

    // drag state
    const [draggingId, setDraggingId] = useState<number | null>(null)
    const [dragOver, setDragOver] = useState<{ parentId: number | null; index: number } | null>(null)

    const load = async () => {
        setLoading(true)
        setError(null)
        try {
            const data = await fetchMenus()
            setTree(data)
        } catch (err: any) {
            setError(err.message || String(err))
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => { load() }, [])

    const createMenu = async (title: string, parent_id?: number) => {
        await fetch('/api/menus', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ title, parent_id }) })
        await load()
    }

    const updateMenu = async (id: number, title: string) => {
        await fetch(`/api/menus/${id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ title }) })
        await load()
    }

    const deleteMenu = async (id: number) => {
        if (!confirm('Delete this item and all children?')) return
        await fetch(`/api/menus/${id}`, { method: 'DELETE' })
        await load()
    }

    // move/reorder API helpers
    const reorderMenu = async (id: number, newOrder: number) => {
        await fetch(`/api/menus/${id}/reorder`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ new_order: newOrder }) })
        await load()
    }

    const moveMenu = async (id: number, newParentId: number | null, newOrder?: number) => {
        await fetch(`/api/menus/${id}/move`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ new_parent_id: newParentId, new_order: newOrder }) })
        await load()
    }

    const filterTree = (nodes: MenuNode[], q: string): MenuNode[] => {
        if (!q) return nodes
        const normalized = q.toLowerCase()
        const out: MenuNode[] = []
        for (const n of nodes) {
            const children = n.children ? filterTree(n.children, q) : []
            const selfMatch = n.title.toLowerCase().includes(normalized)
            if (selfMatch || children.length) out.push({ ...n, children })
        }
        return out
    }

    // helpers to work with the in-memory tree
    const findNode = (nodes: MenuNode[], id: number): MenuNode | null => {
        for (const n of nodes) {
            if (n.id === id) return n
            if (n.children) {
                const r = findNode(n.children, id)
                if (r) return r
            }
        }
        return null
    }

    const isDescendant = (nodes: MenuNode[], ancestorId: number, candidateId: number): boolean => {
        const ancestor = findNode(nodes, ancestorId)
        if (!ancestor || !ancestor.children) return false
        const stack = [...ancestor.children]
        for (const s of stack) {
            if (s.id === candidateId) return true
            if (s.children) stack.push(...s.children)
        }
        return false
    }

    const handleDropAt = async (e: React.DragEvent, targetParentId: number | null, index: number) => {
        e.preventDefault()
        setDragOver(null)
        const raw = e.dataTransfer.getData('application/json')
        if (!raw) return
        let payload: { id: number; parent_id?: number | null; order?: number } = JSON.parse(raw)
        if (!payload || !payload.id) return
        const draggedId = payload.id
        // prevent dropping onto self or descendant
        if (targetParentId !== null && (draggedId === targetParentId || isDescendant(tree, draggedId, targetParentId))) {
            alert('Cannot move an item into itself or its descendant')
            return
        }
        const srcParent = payload.parent_id ?? null
        if (srcParent === targetParentId) {
            // reorder within same parent
            await reorderMenu(draggedId, index)
        } else {
            // move to different parent at index
            await moveMenu(draggedId, targetParentId, index)
        }
    }

    const Node: React.FC<{ node: MenuNode }> = ({ node }) => {
        const [open, setOpen] = useState(true)

        const onDragStart = (e: React.DragEvent) => {
            e.dataTransfer.setData('application/json', JSON.stringify({ id: node.id, parent_id: node.parent_id ?? null, order: node.order }))
            e.dataTransfer.effectAllowed = 'move'
            setDraggingId(node.id)
        }
        const onDragEnd = () => setDraggingId(null)

        const onDropOnNode = async (e: React.DragEvent) => {
            e.preventDefault()
            setDragOver(null)
            const raw = e.dataTransfer.getData('application/json')
            if (!raw) return
            const payload = JSON.parse(raw) as { id: number }
            if (!payload || payload.id === node.id) return
            // prevent moving into descendant
            if (isDescendant(tree, payload.id, node.id)) {
                alert('Cannot move an item into one of its descendants')
                return
            }
            // append as last child
            const idx = (node.children?.length ?? 0)
            await moveMenu(payload.id, node.id, idx)
        }

        return (
            <div className="pl-4 py-1">
                <div className={`flex items-center gap-3 ${draggingId === node.id ? 'opacity-60' : ''}`} draggable onDragStart={onDragStart} onDragEnd={onDragEnd} onDrop={onDropOnNode} onDragOver={e => e.preventDefault()}>
                    <button className="text-sm text-slate-400 mr-2" onClick={() => setOpen(!open)}>
                        {node.children && node.children.length ? (open ? '▾' : '▸') : ''}
                    </button>
                    <div className="flex-1">
                        <span className="font-medium">{node.title}</span>
                        {node.url ? <span className="ml-2 text-xs text-slate-400">({node.url})</span> : null}
                    </div>
                    <div className="flex gap-2">
                        <button className="text-xs px-2 py-0.5 bg-slate-100 rounded" onClick={async (ev) => { ev.stopPropagation(); const title = prompt('Child title'); if (title) await createMenu(title, node.id) }}>Add child</button>
                        <button className="text-xs px-2 py-0.5 bg-yellow-100 rounded" onClick={async (ev) => { ev.stopPropagation(); const title = prompt('New title', node.title); if (title) await updateMenu(node.id, title) }}>Edit</button>
                        <button className="text-xs px-2 py-0.5 bg-rose-100 rounded" onClick={(ev) => { ev.stopPropagation(); deleteMenu(node.id) }}>Delete</button>
                    </div>
                </div>

                <div className={`ml-6 ${open ? 'block' : 'hidden'}`}>
                    {/** child drop-zone before each child and at end */}
                    {(node.children && node.children.length)
                        ? (
                            <div>
                                {node.children!.map((ch, idx) => (
                                    <div key={`wrap-${ch.id}`}>
                                        <div className={`h-2 my-1 rounded ${dragOver && dragOver.parentId === node.id && dragOver.index === idx ? 'bg-sky-200' : 'bg-transparent'}`} onDragOver={e => { e.preventDefault(); setDragOver({ parentId: node.id, index: idx }) }} onDragEnter={e => { e.preventDefault(); setDragOver({ parentId: node.id, index: idx }) }} onDragLeave={() => setDragOver(null)} onDrop={e => handleDropAt(e, node.id, idx)} />
                                        <Node key={ch.id} node={ch} />
                                    </div>
                                ))}
                                <div className={`h-2 my-1 rounded ${dragOver && dragOver.parentId === node.id && dragOver.index === node.children!.length ? 'bg-sky-200' : 'bg-transparent'}`} onDragOver={e => { e.preventDefault(); setDragOver({ parentId: node.id, index: node.children!.length }) }} onDrop={e => handleDropAt(e, node.id, node.children!.length)} />
                            </div>
                        ) : (
                            <div>
                                <div className={`h-2 my-1 rounded ${dragOver && dragOver.parentId === node.id && dragOver.index === 0 ? 'bg-sky-200' : 'bg-transparent'}`} onDragOver={e => { e.preventDefault(); setDragOver({ parentId: node.id, index: 0 }) }} onDrop={e => handleDropAt(e, node.id, 0)} />
                            </div>
                        )}
                </div>
            </div>
        )
    }

    const renderList = (nodes: MenuNode[], parentId: number | null) => {
        const out: React.ReactNode[] = []
        for (let i = 0; i < nodes.length; i++) {
            out.push(<div key={`dz-ro-${parentId ?? 'r'}-${i}`} className={`h-2 my-1 rounded ${dragOver && dragOver.parentId === parentId && dragOver.index === i ? 'bg-sky-200' : 'bg-transparent'}`} onDragOver={e => { e.preventDefault(); setDragOver({ parentId, index: i }) }} onDrop={e => handleDropAt(e, parentId, i)} />)
            out.push(<Node key={nodes[i].id} node={nodes[i]} />)
        }
        out.push(<div key={`dz-ro-${parentId ?? 'r'}-end`} className={`h-2 my-1 rounded ${dragOver && dragOver.parentId === parentId && dragOver.index === nodes.length ? 'bg-sky-200' : 'bg-transparent'}`} onDragOver={e => { e.preventDefault(); setDragOver({ parentId, index: nodes.length }) }} onDrop={e => handleDropAt(e, parentId, nodes.length)} />)
        return out
    }

    const shown = filterTree(tree, search)

    return (
        <div className="max-w-5xl mx-auto p-6">
            <header className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-semibold">Menu Tree — Sotekre</h1>
                <div>
                    <input value={search} onChange={e => setSearch(e.target.value)} className="border px-2 py-1 rounded mr-2" placeholder="Search..." />
                    <button onClick={load} className="bg-sky-600 text-white px-3 py-1 rounded">Refresh</button>
                </div>
            </header>

            <section className="mb-6">
                <form className="flex gap-2" onSubmit={async e => { e.preventDefault(); if (titleInput.trim()) { await createMenu(titleInput.trim()); setTitleInput('') } }}>
                    <input value={titleInput} onChange={e => setTitleInput(e.target.value)} className="flex-1 border px-2 py-1 rounded" placeholder="New root menu title" />
                    <button className="bg-emerald-600 text-white px-3 py-1 rounded">Add root</button>
                </form>
            </section>

            <main>
                <div className="bg-white border rounded p-4 shadow-sm min-h-[200px]">
                    {loading ? <div className="text-sm text-slate-500">Loading…</div> : shown.length === 0 ? <div className="text-sm text-slate-500">No menu items yet</div> : renderList(shown, null)}
                    {error ? <div className="mt-2 text-rose-600 text-sm">{error}</div> : null}
                </div>
            </main>

            <footer className="mt-6 text-sm text-slate-500">Tip: click "Add child" on any item to add nested menus.</footer>
        </div>
    )
}
