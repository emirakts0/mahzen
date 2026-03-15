import * as React from "react"
import { DayPicker } from "react-day-picker"

import { cn } from "@/lib/utils"
import { buttonVariants } from "@/components/ui/button"

export type CalendarProps = React.ComponentProps<typeof DayPicker>

function Calendar({ className, showOutsideDays = true, ...props }: CalendarProps) {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn("p-3", className)}
      style={
        {
          "--cell-width": "2rem",
          "--cell-height": "2rem",
          "--rdp-accent-color": "var(--glass-accent, var(--color-primary))",
          "--rdp-background-color": "var(--glass-hover)",
        } as React.CSSProperties
      }
      classNames={{
        months: "flex flex-col sm:flex-row gap-2",
        month: "flex flex-col gap-4",
        month_caption: "flex justify-center items-center w-full relative py-1",
        caption_label: "text-sm font-medium pointer-events-none",
        nav: "absolute inset-x-0 flex justify-between items-center px-2 z-10",
        button_previous: cn(
          buttonVariants({ variant: "outline" }),
          "h-8 w-8 bg-transparent p-0 hover:opacity-100 cursor-pointer border-0 text-[var(--glass-text)]"
        ),
        button_next: cn(
          buttonVariants({ variant: "outline" }),
          "h-8 w-8 bg-transparent p-0 hover:opacity-100 cursor-pointer border-0 text-[var(--glass-text)]"
        ),
        month_grid: "w-full border-collapse space-x-1",
        weekdays: "flex",
        weekday: "rounded-md w-8 font-normal text-[0.8rem] text-[var(--glass-text-muted)]",
        week: "flex w-full mt-2",
        day: cn(
          "relative p-0 text-center text-sm focus-within:relative focus-within:z-20 [&:first-child:has([aria-selected])]:rounded-l-md [&:last-child:has([aria-selected])]:rounded-r-md",
          "[&:has([aria-selected].day-range-end)]:rounded-r-md [&:has([aria-selected].day-outside)]:bg-accent/50 [&:has([aria-selected])]:bg-accent first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md"
        ),
        day_button: cn(
          buttonVariants({ variant: "ghost" }),
          "h-8 w-8 p-0 font-normal aria-selected:opacity-100 text-[var(--glass-text)]"
        ),
        range_start: "day-range-start rounded-l-md",
        range_end: "day-range-end rounded-r-md",
        selected:
          "bg-[var(--color-primary)] text-[var(--color-primary-foreground)] hover:bg-[var(--color-primary)] hover:text-[var(--color-primary-foreground)] focus:bg-[var(--color-primary)] focus:text-[var(--color-primary-foreground)]",
        today: "bg-[var(--glass-hover)] text-[var(--glass-text)]",
        outside:
          "day-outside text-[var(--glass-text-muted)] opacity-50 aria-selected:bg-[var(--glass-hover)] aria-selected:opacity-30",
        disabled: "opacity-50",
        range_middle:
          "aria-selected:bg-[var(--glass-hover)] aria-selected:text-[var(--glass-text)]",
        hidden: "invisible",
      }}
      {...props}
    />
  )
}
Calendar.displayName = "Calendar"

export { Calendar }
