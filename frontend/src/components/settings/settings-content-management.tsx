import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { FileText, Play } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useI18n } from "@/lib/i18n";
import { APIError } from "@/lib/api";
import {
  fullContentAPI,
  type FullContentBackfillStatus,
} from "@/lib/api/fullcontent.ext";

const INTERVAL_STORAGE_KEY = "fusion-backfill-interval-seconds";
const DAYS_STORAGE_KEY = "fusion-backfill-days";
const LIMIT_STORAGE_KEY = "fusion-backfill-limit";

const DEFAULT_INTERVAL = 1;
const DEFAULT_DAYS = 30;
const DEFAULT_LIMIT = 500;

function readStoredNumber(key: string, fallback: number): number {
  const raw = localStorage.getItem(key);
  if (raw === null || raw.trim() === "") return fallback;
  const parsed = Number.parseInt(raw, 10);
  return Number.isNaN(parsed) ? fallback : parsed;
}

function NumericField({
  label,
  description,
  value,
  min,
  max,
  onChange,
}: {
  label: string;
  description: string;
  value: number;
  min: number;
  max: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="flex items-center justify-between gap-6">
      <div className="min-w-0 space-y-1">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-[13px] text-muted-foreground">{description}</p>
      </div>
      <Input
        type="number"
        className="w-24"
        min={min}
        max={max}
        value={Number.isNaN(value) ? "" : value}
        onChange={(e) => {
          const parsed = Number.parseInt(e.target.value, 10);
          onChange(Number.isNaN(parsed) ? NaN : parsed);
        }}
      />
    </div>
  );
}

export function ContentManagement() {
  const { t } = useI18n();

  const [intervalSeconds, setIntervalSeconds] = useState(() =>
    readStoredNumber(INTERVAL_STORAGE_KEY, DEFAULT_INTERVAL),
  );
  const [days, setDays] = useState(() =>
    readStoredNumber(DAYS_STORAGE_KEY, DEFAULT_DAYS),
  );
  const [limit, setLimit] = useState(() =>
    readStoredNumber(LIMIT_STORAGE_KEY, DEFAULT_LIMIT),
  );

  const statusQuery = useQuery<FullContentBackfillStatus>({
    queryKey: ["fullContentBackfillStatus"],
    queryFn: async () => {
      const res = await fullContentAPI.getBackfillStatus();
      return res.data!;
    },
    refetchInterval: (query) => (query.state.data?.running ? 2000 : false),
  });
  const status = statusQuery.data;

  const startMutation = useMutation({
    mutationFn: async () => {
      const res = await fullContentAPI.startBackfill({
        days: Number.isNaN(days) ? DEFAULT_DAYS : Math.max(0, days),
        limit:
          Number.isNaN(limit) || limit <= 0 ? DEFAULT_LIMIT : Math.min(limit, 10000),
        interval_seconds: Number.isNaN(intervalSeconds)
          ? DEFAULT_INTERVAL
          : Math.min(Math.max(intervalSeconds, 0), 60),
      });
      return res.data!;
    },
    onSuccess: () => {
      toast.success(t("settings.contentManagement.toast.started"));
      void statusQuery.refetch();
    },
    onError: (error) => {
      if (error instanceof APIError && error.status === 409) {
        toast.error(t("settings.contentManagement.toast.alreadyRunning"));
      } else {
        toast.error(t("settings.contentManagement.toast.startFailed"));
      }
    },
  });

  const running = status?.running ?? false;
  const hasLastRun =
    !running && !!status && (status.done > 0 || status.failed > 0 || status.skipped > 0);

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between gap-6">
        <div className="min-w-0 space-y-1">
          <p className="flex items-center gap-2 text-sm font-medium">
            <FileText className="h-4 w-4" />
            {t("settings.contentManagement.backfillLabel")}
          </p>
          <p className="text-[13px] text-muted-foreground">
            {t("settings.contentManagement.backfillDescription")}
          </p>
        </div>
      </div>

      <NumericField
        label={t("settings.contentManagement.interval.label")}
        description={t("settings.contentManagement.interval.description")}
        value={intervalSeconds}
        min={0}
        max={60}
        onChange={(value) => {
          setIntervalSeconds(value);
          localStorage.setItem(INTERVAL_STORAGE_KEY, String(value));
        }}
      />

      <NumericField
        label={t("settings.contentManagement.days.label")}
        description={t("settings.contentManagement.days.description")}
        value={days}
        min={0}
        max={3650}
        onChange={(value) => {
          setDays(value);
          localStorage.setItem(DAYS_STORAGE_KEY, String(value));
        }}
      />

      <NumericField
        label={t("settings.contentManagement.limit.label")}
        description={t("settings.contentManagement.limit.description")}
        value={limit}
        min={0}
        max={10000}
        onChange={(value) => {
          setLimit(value);
          localStorage.setItem(LIMIT_STORAGE_KEY, String(value));
        }}
      />

      <div className="flex items-center justify-between gap-6">
        <div className="min-w-0 space-y-1">
          {running && status && (
            <>
              <p className="text-sm font-medium">
                {t("settings.contentManagement.running")}
              </p>
              <p className="text-[13px] text-muted-foreground">
                {t("settings.contentManagement.counters", {
                  done: status.done,
                  failed: status.failed,
                  skipped: status.skipped,
                })}
              </p>
            </>
          )}
          {hasLastRun && status && (
            <p className="text-sm text-muted-foreground">
              {t("settings.contentManagement.lastRun", {
                done: status.done,
                failed: status.failed,
                skipped: status.skipped,
              })}
            </p>
          )}
        </div>
        <Button
          size="sm"
          disabled={running || startMutation.isPending}
          onClick={() => startMutation.mutate()}
        >
          <Play className="h-4 w-4" />
          {t("settings.contentManagement.start")}
        </Button>
      </div>
    </div>
  );
}
