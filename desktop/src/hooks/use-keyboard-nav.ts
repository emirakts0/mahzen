import { useEffect, useState, useCallback, useRef } from "react";
import { writeText } from "@tauri-apps/plugin-clipboard-manager";
import { getCurrentWindow } from "@tauri-apps/api/window";
import type { SearchResult } from "@/types/api";

export type SearchColumn = "keyword" | "semantic";

interface UseKeyboardNavOptions {
  keywordResults: SearchResult[];
  semanticResults: SearchResult[];
  onPreview?: (result: SearchResult) => void;
  onColumnChange?: (column: SearchColumn) => void;
  searchInputRef?: React.RefObject<HTMLInputElement | null>;
}

export function useKeyboardNav({
  keywordResults,
  semanticResults,
  onPreview,
  onColumnChange,
  searchInputRef,
}: UseKeyboardNavOptions) {
  const [activeColumn, setActiveColumn] = useState<SearchColumn>("keyword");
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const selectedRef = useRef<HTMLDivElement>(null);

  const currentResults = activeColumn === "keyword" ? keywordResults : semanticResults;

  const handleCopy = useCallback(async (result: SearchResult) => {
    try {
      await writeText(result.content ?? result.summary ?? "");
    } catch (error) {
      console.error("Failed to copy:", error);
    }
  }, []);

  const handleColumnChange = useCallback((column: SearchColumn) => {
    setActiveColumn(column);
    setSelectedIndex(-1);
    onColumnChange?.(column);
  }, [onColumnChange]);

  useEffect(() => {
    if (selectedRef.current) {
      selectedRef.current.scrollIntoView({ behavior: "smooth", block: "nearest" });
    }
  }, [selectedIndex]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const isInputFocused = document.activeElement === searchInputRef?.current;
      
      if (e.key === "Escape") {
        e.preventDefault();
        if (selectedIndex >= 0) {
          setSelectedIndex(-1);
        } else {
          void getCurrentWindow().hide();
        }
        return;
      }

      if (e.key === "ArrowDown") {
        e.preventDefault();
        if (isInputFocused) {
          // Move from input to first result
          setSelectedIndex(0);
          searchInputRef?.current?.blur();
        } else {
          setSelectedIndex((i) => Math.min(i + 1, currentResults.length - 1));
        }
        return;
      }

      if (e.key === "ArrowUp" && !isInputFocused) {
        e.preventDefault();
        if (selectedIndex === 0) {
          // Move back to input
          setSelectedIndex(-1);
          searchInputRef?.current?.focus();
        } else {
          setSelectedIndex((i) => Math.max(i - 1, 0));
        }
        return;
      }

      if (e.key === "ArrowLeft") {
        e.preventDefault();
        if (!isInputFocused) {
          handleColumnChange("keyword");
        }
        return;
      }

      if (e.key === "ArrowRight") {
        e.preventDefault();
        if (!isInputFocused) {
          handleColumnChange("semantic");
        } else if (currentResults.length > 0) {
          // Move from input to results when pressing right
          handleColumnChange("semantic");
          setSelectedIndex(0);
          searchInputRef?.current?.blur();
        }
        return;
      }

      if (currentResults.length === 0) return;

      if (e.key === "c" && e.ctrlKey && !e.shiftKey) {
        e.preventDefault();
        const selected = currentResults[selectedIndex];
        if (selected) {
          void handleCopy(selected);
        }
      } else if (e.key === "Enter" && !isInputFocused) {
        e.preventDefault();
        const selected = currentResults[selectedIndex];
        if (selected && onPreview) {
          onPreview(selected);
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [currentResults, selectedIndex, onPreview, handleCopy, handleColumnChange, searchInputRef]);

  useEffect(() => {
    setSelectedIndex(-1);
  }, [keywordResults, semanticResults, activeColumn]);

  return { 
    selectedIndex, 
    setSelectedIndex, 
    activeColumn, 
    setActiveColumn: handleColumnChange,
    selectedRef,
  };
}
