import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api"
import type { CreateTagRequest } from "@/lib/types"

export function useTags(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ["tags", params],
    queryFn: () => api.listTags(params),
  })
}

export function useTag(id: string) {
  return useQuery({
    queryKey: ["tags", id],
    queryFn: () => api.getTag(id),
    enabled: !!id,
  })
}

export function useCreateTag() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateTagRequest) => api.createTag(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tags"] })
    },
  })
}

export function useDeleteTag() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteTag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tags"] })
    },
  })
}

export function useAttachTag() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ entryId, tagId }: { entryId: string; tagId: string }) =>
      api.attachTag(entryId, { tag_id: tagId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["entries"] })
    },
  })
}

export function useDetachTag() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ entryId, tagId }: { entryId: string; tagId: string }) =>
      api.detachTag(entryId, tagId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["entries"] })
    },
  })
}
