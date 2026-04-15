import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { itemAPI } from "@/lib/api";
import type { Item } from "@/lib/api";
import { queryKeys } from "@/queries/keys";

export function useAutoMarkRead(article: Item | null, canToggleRead: boolean) {
  const qc = useQueryClient();
  const markedRef = useRef<Set<number>>(new Set());

  useEffect(() => {
    if (!article || !article.unread || !canToggleRead) return;
    if (markedRef.current.has(article.id)) return;

    markedRef.current.add(article.id);
    const timer = window.setTimeout(async () => {
      try {
        await itemAPI.markRead({ ids: [article.id] });

        qc.setQueriesData<{
          pages: Array<{ data: Item[]; total: number }>;
        }>({ queryKey: queryKeys.items.lists() }, (old) => {
          if (!old?.pages) return old;
          return {
            ...old,
            pages: old.pages.map((page) => ({
              ...page,
              data: page.data.map((item) =>
                item.id === article.id ? { ...item, unread: false } : item,
              ),
            })),
          };
        });

        // Update detail cache
        qc.setQueryData<Item>(queryKeys.items.detail(article.id), (old) =>
          old ? { ...old, unread: false } : old,
        );

        qc.setQueryData<Array<{ id: number; unread_count: number }>>(
          queryKeys.feeds.list(),
          (old) =>
            old?.map((feed) =>
              feed.id === article.feed_id
                ? { ...feed, unread_count: Math.max(0, feed.unread_count - 1) }
                : feed,
            ),
        );
      } catch (error) {
        console.error("Failed to auto-mark read:", error);
      }
    }, 300);

    return () => window.clearTimeout(timer);
  }, [article?.id, article?.unread, canToggleRead, qc]);
}
