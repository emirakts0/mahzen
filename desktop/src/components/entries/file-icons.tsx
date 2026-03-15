import {
  FileText,
  FileCode,
  FileVideo,
  FileImage,
  FileArchive,
  FileSpreadsheet,
  FileType,
} from "lucide-react"
import type { LucideIcon } from "lucide-react"

const iconMap: Record<string, LucideIcon> = {
  go: FileCode,
  ts: FileCode,
  tsx: FileCode,
  js: FileCode,
  jsx: FileCode,
  py: FileCode,
  rs: FileCode,
  java: FileCode,
  c: FileCode,
  cpp: FileCode,
  h: FileCode,
  css: FileCode,
  html: FileCode,
  json: FileCode,
  yaml: FileCode,
  yml: FileCode,
  toml: FileCode,
  sh: FileCode,
  bash: FileCode,
  mp4: FileVideo,
  avi: FileVideo,
  mov: FileVideo,
  mkv: FileVideo,
  webm: FileVideo,
  png: FileImage,
  jpg: FileImage,
  jpeg: FileImage,
  gif: FileImage,
  webp: FileImage,
  svg: FileImage,
  ico: FileImage,
  zip: FileArchive,
  tar: FileArchive,
  gz: FileArchive,
  rar: FileArchive,
  "7z": FileArchive,
  bz2: FileArchive,
  xlsx: FileSpreadsheet,
  xls: FileSpreadsheet,
  csv: FileSpreadsheet,
  tsv: FileSpreadsheet,
  pdf: FileType,
  doc: FileType,
  docx: FileType,
  txt: FileType,
  md: FileType,
  rtf: FileType,
}

interface FileIconProps {
  fileType?: string
  className?: string
  style?: React.CSSProperties
}

export function FileIcon({ fileType, className, style }: FileIconProps) {
  const Icon = fileType ? iconMap[fileType.toLowerCase()] ?? FileText : FileText
  return <Icon className={className} style={style} />
}

export function getFileTypeLabel(fileType?: string): string {
  if (!fileType) return "Text"

  const type = fileType.toLowerCase()

  const labels: Record<string, string> = {
    go: "Go",
    ts: "TypeScript",
    tsx: "TypeScript React",
    js: "JavaScript",
    jsx: "JavaScript React",
    py: "Python",
    rs: "Rust",
    java: "Java",
    c: "C",
    cpp: "C++",
    css: "CSS",
    html: "HTML",
    json: "JSON",
    yaml: "YAML",
    yml: "YAML",
    toml: "TOML",
    sh: "Shell",
    bash: "Bash",
    mp4: "MP4 Video",
    avi: "AVI Video",
    mov: "QuickTime",
    mkv: "Matroska",
    webm: "WebM",
    png: "PNG Image",
    jpg: "JPEG Image",
    jpeg: "JPEG Image",
    gif: "GIF Image",
    webp: "WebP Image",
    svg: "SVG Image",
    zip: "ZIP Archive",
    tar: "TAR Archive",
    gz: "GZIP Archive",
    rar: "RAR Archive",
    "7z": "7-Zip Archive",
    xlsx: "Excel Spreadsheet",
    xls: "Excel Spreadsheet",
    csv: "CSV Data",
    pdf: "PDF Document",
    doc: "Word Document",
    docx: "Word Document",
    txt: "Plain Text",
    md: "Markdown",
  }

  return labels[type] || type.toUpperCase()
}

export function formatFileSize(bytes?: number): string {
  if (!bytes) return ""

  const units = ["B", "KB", "MB", "GB"]
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}
