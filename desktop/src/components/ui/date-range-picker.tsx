import { useState, useEffect } from "react"
import { format, subDays, startOfMonth, startOfYear, subMonths } from "date-fns"
import { Calendar as CalendarIcon } from "lucide-react"
import type { DateRange as DayPickerDateRange } from "react-day-picker"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface DateRange {
  from?: string
  to?: string
}

interface DateRangePickerProps {
  value: DateRange
  onChange: (range: DateRange) => void
}

type PresetKey = "today" | "yesterday" | "last7" | "last30" | "thisMonth" | "lastMonth" | "thisYear"

const presets: { key: PresetKey; label: string }[] = [
  { key: "today", label: "Today" },
  { key: "yesterday", label: "Yesterday" },
  { key: "last7", label: "Last 7 days" },
  { key: "last30", label: "Last 30 days" },
  { key: "thisMonth", label: "This month" },
  { key: "lastMonth", label: "Last month" },
  { key: "thisYear", label: "This year" },
]

function getPresetRange(preset: PresetKey): DayPickerDateRange | undefined {
  const today = new Date()

  switch (preset) {
    case "today":
      return { from: today, to: today }
    case "yesterday": {
      const y = subDays(today, 1)
      return { from: y, to: y }
    }
    case "last7":
      return { from: subDays(today, 6), to: today }
    case "last30":
      return { from: subDays(today, 29), to: today }
    case "thisMonth":
      return { from: startOfMonth(today), to: today }
    case "lastMonth": {
      const lastMonth = subMonths(today, 1)
      return { from: startOfMonth(lastMonth), to: today }
    }
    case "thisYear":
      return { from: startOfYear(today), to: today }
  }
}

function detectPreset(range: DateRange): PresetKey | null {
  if (!range.from || !range.to) return null

  for (const preset of presets) {
    const presetRange = getPresetRange(preset.key)
    if (
      presetRange?.from &&
      presetRange?.to &&
      format(presetRange.from, "yyyy-MM-dd") === range.from &&
      format(presetRange.to, "yyyy-MM-dd") === range.to
    ) {
      return preset.key
    }
  }

  return null
}

function dateRangeToString(range: DayPickerDateRange | undefined): DateRange {
  return {
    from: range?.from ? format(range.from, "yyyy-MM-dd") : undefined,
    to: range?.to ? format(range.to, "yyyy-MM-dd") : undefined,
  }
}

function stringToDateRange(range: DateRange): DayPickerDateRange | undefined {
  if (!range.from) return undefined
  return {
    from: new Date(range.from),
    to: range.to ? new Date(range.to) : undefined,
  }
}

export function DateRangePicker({ value, onChange }: DateRangePickerProps) {
  const [selectedPreset, setSelectedPreset] = useState<PresetKey | null>(() => detectPreset(value))
  const [date, setDate] = useState<DayPickerDateRange | undefined>(() => stringToDateRange(value))

  useEffect(() => {
    setSelectedPreset(detectPreset(value))
    setDate(stringToDateRange(value))
  }, [value])

  const handlePresetSelect = (preset: PresetKey) => {
    setSelectedPreset(preset)
    const range = getPresetRange(preset)
    setDate(range)
    onChange(dateRangeToString(range))
  }

  const handleDateSelect = (range: DayPickerDateRange | undefined) => {
    setDate(range)
    setSelectedPreset(null)
    onChange(dateRangeToString(range))
  }

  const displayLabel = () => {
    if (selectedPreset) {
      return presets.find((p) => p.key === selectedPreset)?.label ?? "Select date range"
    }
    if (date?.from) {
      return date.to
        ? `${format(date.from, "MMM d")} - ${format(date.to, "MMM d, yyyy")}`
        : format(date.from, "MMM d, yyyy")
    }
    return "Select date range"
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap gap-1">
        {presets.map((preset) => (
          <button
            key={preset.key}
            type="button"
            onClick={() => handlePresetSelect(preset.key)}
            className="rounded-md px-2 py-1 text-xs font-medium transition-all hover:opacity-80"
            style={{
              background: selectedPreset === preset.key ? "var(--glass-accent, #3b82f6)" : "var(--glass-hover)",
              color: selectedPreset === preset.key ? "white" : "var(--glass-text)",
              border: "1px solid var(--glass-border)",
            }}
          >
            {preset.label}
          </button>
        ))}
      </div>

      <Popover>
        <PopoverTrigger asChild>
          <button
            type="button"
            className="flex h-8 w-full items-center justify-between gap-2 rounded-md px-2 text-sm outline-none backdrop-blur-sm transition-colors"
            style={{
              background: "var(--glass-hover)",
              border: "1px solid var(--glass-border)",
              color: date?.from ? "var(--glass-text)" : "var(--glass-text-muted)",
            }}
          >
            <div className="flex items-center gap-1.5">
              <CalendarIcon className="h-3.5 w-3.5" style={{ color: "var(--glass-text-muted)" }} />
              <span className="truncate">{displayLabel()}</span>
            </div>
          </button>
        </PopoverTrigger>
        <PopoverContent
          className="w-auto p-0"
          align="start"
          side="bottom"
          sideOffset={4}
          avoidCollisions
          collisionPadding={16}
          container={typeof document !== "undefined" ? document.body : undefined}
          style={{
            background: "var(--glass-bg)",
            border: "1px solid var(--glass-border)",
            backdropFilter: "blur(16px)",
            zIndex: 9999,
            color: "var(--glass-text)",
          }}
        >
          <Calendar
            mode="range"
            defaultMonth={date?.from}
            selected={date}
            onSelect={handleDateSelect}
            numberOfMonths={1}
          />
          {date?.from && (
            <div className="border-t p-2" style={{ borderColor: "var(--glass-border)" }}>
              <button
                type="button"
                onClick={() => {
                  setDate(undefined)
                  setSelectedPreset(null)
                  onChange({})
                }}
                className="w-full rounded-md px-2 py-1.5 text-xs font-medium transition-colors"
                style={{
                  background: "var(--glass-hover)",
                  color: "var(--glass-text-muted)",
                }}
              >
                Clear dates
              </button>
            </div>
          )}
        </PopoverContent>
      </Popover>
    </div>
  )
}
