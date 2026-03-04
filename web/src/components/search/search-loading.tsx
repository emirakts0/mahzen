import { motion } from "framer-motion"
import { BookText, Sparkles } from "lucide-react"

function WavyLoadingHeader({
  icon: Icon,
  title,
  delay,
}: {
  icon: React.ElementType
  title: string
  delay: number
}) {
  return (
    <motion.div
      className="flex items-center gap-2 pb-1"
      initial={{ opacity: 0.9 }}
      animate={{
        opacity: [0.9, 0.4, 0.9],
      }}
      transition={{
        duration: 1.5,
        repeat: Infinity,
        ease: "easeInOut",
        delay,
      }}
    >
      <div
        className="flex h-7 w-7 items-center justify-center rounded-lg backdrop-blur-sm"
        style={{ background: "var(--glass-hover)" }}
      >
        <Icon className="h-4 w-4" style={{ color: "var(--glass-text-muted)" }} />
      </div>
      <span
        className="text-sm font-semibold uppercase tracking-widest"
        style={{ color: "var(--glass-text-muted)" }}
      >
        {title}
      </span>
    </motion.div>
  )
}

export function SearchLoading() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
      <div className="flex flex-col gap-3">
        <WavyLoadingHeader icon={BookText} title="Keyword Search" delay={0} />
      </div>
      <div className="flex flex-col gap-3">
        <WavyLoadingHeader icon={Sparkles} title="Semantic Search" delay={0.2} />
      </div>
    </div>
  )
}
