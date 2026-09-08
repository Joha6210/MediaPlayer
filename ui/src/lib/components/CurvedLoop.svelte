<script lang="ts">
  type Direction = 'left' | 'right';

  type Props = {
    marqueeText?: string;
    speed?: number;
    class?: string;
    curveAmount?: number;
    direction?: Direction;
    interactive?: boolean;
  };

  let {
    marqueeText = '',
    speed = 2,
    class: className = '',
    curveAmount = 400,
    direction = 'left',
    interactive = true
  }: Props = $props();

  const text = $derived.by(() => {
    const hasTrailing = /\s|\u00A0$/.test(marqueeText);
    return (hasTrailing ? marqueeText.replace(/\s+$/, '') : marqueeText) + '\u00A0';
  });

  let measureEl: SVGTextElement | undefined = $state();
  let textPathEl: SVGTextPathElement | undefined = $state();
  let spacing = $state(0);
  let offset = $state(0);
  let isDragging = $state(false);

  const uid = $props.id();
  const pathId = `curve-${uid}`;
  const pathD = $derived(`M-100,40 Q500,${40 + curveAmount} 1540,40`);

  let dragActive = false;
  let lastX = 0;
  let velX = 0;
  // svelte-ignore state_referenced_locally
  let dirInternal: Direction = direction;

  $effect(() => {
    dirInternal = direction;
  });

  const totalText = $derived.by(() => {
    if (!spacing) return text;
    return Array(Math.ceil(1800 / spacing) + 2).fill(text).join('');
  });
  const ready = $derived(spacing > 0);

  $effect(() => {
    void text;
    void className;
    if (measureEl) {
      spacing = measureEl.getComputedTextLength();
    }
  });

  $effect(() => {
    if (!spacing || !textPathEl) return;
    const initial = 15;
    textPathEl.setAttribute('startOffset', `${initial}px`);
    offset = initial;
  });

  $effect(() => {
    if (!spacing || !ready) return;
    void speed;
    let frame = 0;
    const initialOffset = 15;
    const pauseDuration = 10000; // 10 seconds
    let pauseUntil = 0;

    const step = (timestamp: number) => {
      if (timestamp >= pauseUntil && !dragActive && textPathEl) {
        const delta = dirInternal === 'right' ? speed : -speed;
        const currentOffset = parseFloat(textPathEl.getAttribute('startOffset') || '0');
        let newOffset = currentOffset + delta;

        if (dirInternal === 'left' && newOffset <= initialOffset - spacing) {
          newOffset = initialOffset;
          pauseUntil = timestamp + pauseDuration;
        } else if (dirInternal === 'right' && newOffset >= initialOffset + spacing) {
          newOffset = initialOffset;
          pauseUntil = timestamp + pauseDuration;
        }

        textPathEl.setAttribute('startOffset', `${newOffset}px`);
        offset = newOffset;
      }
      frame = requestAnimationFrame(step);
    };
    frame = requestAnimationFrame(step);
    return () => cancelAnimationFrame(frame);
  });

  function onPointerDown(event: PointerEvent) {
    if (!interactive) return;
    dragActive = true;
    isDragging = true;
    lastX = event.clientX;
    velX = 0;
    (event.target as Element)?.setPointerCapture?.(event.pointerId);
  }

  function onPointerMove(event: PointerEvent) {
    if (!interactive || !dragActive || !textPathEl) return;
    const deltaX = event.clientX - lastX;
    lastX = event.clientX;
    velX = deltaX;

    const currentOffset = parseFloat(textPathEl.getAttribute('startOffset') || '0');
    let newOffset = currentOffset + deltaX;

    if (newOffset <= -spacing) newOffset += spacing;
    if (newOffset > 0) newOffset -= spacing;

    textPathEl.setAttribute('startOffset', `${newOffset}px`);
    offset = newOffset;
  }

  function endDrag() {
    if (!interactive) return;
    dragActive = false;
    isDragging = false;
    dirInternal = velX > 0 ? 'right' : 'left';
  }

  const cursorStyle = $derived(interactive ? (isDragging ? 'grabbing' : 'grab') : 'auto');
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="curved-loop-jacket"
  style:visibility={ready ? 'visible' : 'hidden'}
  style:cursor={cursorStyle}
  onpointerdown={onPointerDown}
  onpointermove={onPointerMove}
  onpointerup={endDrag}
  onpointerleave={endDrag}
>
  <svg class="curved-loop-svg" viewBox="0 -60 1440 180">
    <text bind:this={measureEl} xml:space="preserve" style="visibility:hidden;opacity:0;pointer-events:none;">{text}</text>
    <defs>
      <path id={pathId} d={pathD} fill="none" stroke="transparent" />
    </defs>
    {#if ready}
      <text font-weight="bold" xml:space="preserve" class={className}>
        <textPath bind:this={textPathEl} href={`#${pathId}`} startOffset={`${offset}px`} xml:space="preserve">{totalText}</textPath>
      </text>
    {/if}
  </svg>
</div>

<style>
  .curved-loop-jacket {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 96px;
    overflow: hidden;
  }

  .curved-loop-svg {
    display: block;
    width: 100%;
    aspect-ratio: 100 / 12;
    overflow: visible;
    user-select: none;
    font-size: 2rem;
    fill: #ffffff;
    font-weight: 700;
    text-transform: uppercase;
  }
</style>