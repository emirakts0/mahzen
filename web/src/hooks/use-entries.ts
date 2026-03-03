import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api"
import type {
  CreateEntryRequest,
  UpdateEntryRequest,
} from "@/lib/types"

export function useEntries(params?: {
  path?: string
  limit?: number
  offset?: number
}) {
  return useQuery({
    queryKey: ["entries", params],
    queryFn: () => api.listEntries(params),
  })
}

export function useEntry(id: string) {
  return useQuery({
    queryKey: ["entries", id],
    queryFn: () => api.getEntry(id),
    enabled: !!id,
  })
}

export function useCreateEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateEntryRequest) => api.createEntry(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["entries"] })
    },
  })
}

export function useUpdateEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateEntryRequest }) =>
      api.updateEntry(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["entries"] })
      queryClient.invalidateQueries({ queryKey: ["entries", variables.id] })
    },
  })
}

export function useDeleteEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["entries"] })
    },
  })
}
