import axios from 'axios'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export interface MenuNode {
    id: number
    title: string
    url?: string
    parent_id?: number
    order: number
    children?: MenuNode[]
}

export interface MenuResponse {
    data: MenuNode[]
}

export interface CreateMenuInput {
    title: string
    url?: string
    parent_id?: number
    order?: number
}

export interface UpdateMenuInput {
    title?: string
    url?: string
    parent_id?: number
    order?: number
}

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
})

export const menuService = {
    // Get all menus
    async getMenus(): Promise<MenuNode[]> {
        // Add cache-busting parameter to ensure fresh data
        const timestamp = new Date().getTime()
        const response = await api.get<MenuResponse>(`/api/menus?_t=${timestamp}`)
        return response.data.data || []
    },

    // Create menu
    async createMenu(input: CreateMenuInput): Promise<MenuNode> {
        const response = await api.post<MenuNode>('/api/menus', input)
        return response.data
    },

    // Update menu
    async updateMenu(id: number, input: UpdateMenuInput): Promise<MenuNode> {
        const response = await api.put<MenuNode>(`/api/menus/${id}`, input)
        return response.data
    },

    // Delete menu
    async deleteMenu(id: number): Promise<void> {
        await api.delete(`/api/menus/${id}`)
    },

    // Reorder menu
    async reorderMenu(id: number, newOrder: number): Promise<void> {
        await api.patch(`/api/menus/${id}/reorder`, { new_order: newOrder })
    },

    // Move menu
    async moveMenu(id: number, newParentId: number | null, newOrder?: number): Promise<void> {
        await api.patch(`/api/menus/${id}/move`, {
            new_parent_id: newParentId,
            new_order: newOrder,
        })
    },
}
