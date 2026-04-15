import { useMutation, useQueryClient } from "@tanstack/react-query";
import { groupAPI, type Group } from "@/lib/api";
import { queryKeys } from "./keys";
import { toast } from "sonner";
import type { TranslationKey } from "@/lib/i18n";

export function useUpdateGroupExtended() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      name,
      auto_fetch_full_content,
    }: {
      id: number;
      name?: string;
      auto_fetch_full_content?: boolean | null;
    }) => {
      await groupAPI.update(id, { name, auto_fetch_full_content });
      return { id, name, auto_fetch_full_content };
    },
    onSuccess: ({ id, name, auto_fetch_full_content }) => {
      qc.setQueryData(queryKeys.groups.list(), (old: Group[] | undefined) =>
        old?.map((g) =>
          g.id === id
            ? {
                ...g,
                ...(name !== undefined && { name }),
                ...(auto_fetch_full_content !== undefined && {
                  auto_fetch_full_content,
                }),
              }
            : g,
        ),
      );
    },
  });
}

export function useCreateGroupExtended() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      name: string;
      auto_fetch_full_content?: boolean | null;
    }) => {
      const res = await groupAPI.create(payload);
      return res.data!;
    },
    onSuccess: (group) => {
      qc.setQueryData(queryKeys.groups.list(), (old: Group[] | undefined) =>
        old ? [...old, group] : [group],
      );
    },
  });
}

export function useSaveGroupExtended(
  t: (key: TranslationKey) => string,
) {
  const updateGroupMutation = useUpdateGroupExtended();

  return {
    isPending: updateGroupMutation.isPending,
    save: async (
      group: Group,
      editingName: string,
      autoFetch?: boolean | null,
    ) => {
      const name = editingName.trim();
      const shouldUpdateName = name && name !== group.name;
      const shouldUpdateAutoFetch =
        autoFetch !== undefined && autoFetch !== group.auto_fetch_full_content;

      if (!shouldUpdateName && !shouldUpdateAutoFetch) return;

      try {
        await updateGroupMutation.mutateAsync({
          id: group.id,
          ...(shouldUpdateName && { name }),
          ...(shouldUpdateAutoFetch && { auto_fetch_full_content: autoFetch }),
        });
        toast.success(
          t(
            shouldUpdateName
              ? "feeds.toast.renamed"
              : "group.toast.updated",
          ),
        );
      } catch {
        toast.error(
          t(
            shouldUpdateName
              ? "feeds.toast.renameFailed"
              : "group.toast.updateFailed",
          ),
        );
      }
    },
  };
}
