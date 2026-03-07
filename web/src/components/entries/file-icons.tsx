import {
  FileText,
  FileCode,
  FileVideo,
  FileImage,
  FileArchive,
  FileSpreadsheet,
  FileType,
} from "lucide-react"

export function getFileIcon(fileType?: string) {
  if (!fileType) return FileText

  const type = fileType.toLowerCase()

  switch (type) {
    case "go":
    case "ts":
    case "tsx":
    case "js":
    case "jsx":
    case "py":
    case "rs":
    case "java":
    case "c":
    case "cpp":
    case "h":
    case "css":
    case "html":
    case "json":
    case "yaml":
    case "yml":
    case "toml":
    case "sh":
    case "bash":
      return FileCode

    case "mp4":
    case "avi":
    case "mov":
    case "mkv":
    case "webm":
      return FileVideo

    case "png":
    case "jpg":
    case "jpeg":
    case "gif":
    case "webp":
    case "svg":
    case "ico":
      return FileImage

    case "zip":
    case "tar":
    case "gz":
    case "rar":
    case "7z":
    case "bz2":
      return FileArchive

    case "xlsx":
    case "xls":
    case "csv":
    case "tsv":
      return FileSpreadsheet

    case "pdf":
    case "doc":
    case "docx":
    case "txt":
    case "md":
    case "rtf":
      return FileType

    default:
      return FileText
  }
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
