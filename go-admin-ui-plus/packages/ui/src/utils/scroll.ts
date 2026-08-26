const easeInOutQuad = (time: number, start: number, change: number, duration: number) => {
  let progress = time / (duration / 2)
  if (progress < 1) return change / 2 * progress * progress + start
  progress--
  return -change / 2 * (progress * (progress - 2) - 1) + start
}

const position = () => document.documentElement.scrollTop || document.body.scrollTop

export function scrollTo(target: number, duration = 500, callback?: () => void) {
  const start = position()
  const change = target - start
  let elapsed = 0
  const animate = () => {
    elapsed += 20
    const next = easeInOutQuad(elapsed, start, change, duration)
    document.documentElement.scrollTop = next
    document.body.scrollTop = next
    if (elapsed < duration) window.requestAnimationFrame(animate)
    else callback?.()
  }
  animate()
}
