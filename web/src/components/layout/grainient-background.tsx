import Grainient from "@/components/Grainient"
import { useTheme } from "@/hooks/use-theme"

// Dark mode: deep navy blues + bright red accent
const DARK_PALETTE = {
  color1: "#050E3C",
  color2: "#002455",
  color3: "#FF3838",
} as const

// Light mode: pastel versions of the same palette
const LIGHT_PALETTE = {
  color1: "#8fa3d4",
  color2: "#6b94c8",
  color3: "#f08080",
} as const

export default function GrainientBackground() {
  const { theme, bgAnimation } = useTheme()
  const palette = theme === "dark" ? DARK_PALETTE : LIGHT_PALETTE

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 0,
        width: "100vw",
        height: "100vh",
      }}
    >
      <Grainient
        animated={bgAnimation}
        color1={palette.color1}
        color2={palette.color2}
        color3={palette.color3}
        timeSpeed={0.75}
        colorBalance={0}
        warpStrength={1}
        warpFrequency={5}
        warpSpeed={2}
        warpAmplitude={50}
        blendAngle={0}
        blendSoftness={0.05}
        rotationAmount={500}
        noiseScale={2}
        grainAmount={0.1}
        grainScale={2}
        grainAnimated={false}
        contrast={1.0}
        gamma={1.1}
        saturation={1.1}
        centerX={0}
        centerY={0}
        zoom={0.9}
      />
    </div>
  )
}
