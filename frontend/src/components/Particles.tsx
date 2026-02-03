import { useEffect, useRef } from 'react';

interface Particle {
  x: number;
  y: number;
  vx: number;
  vy: number;
  baseVx: number;
  baseVy: number;
  radius: number;
  baseRadius: number; // Original radius for scaling
  color: string;
  alpha: number;
  trail: { x: number; y: number; alpha: number }[];
  waveOffset: number;
  waveAmplitude: number;
  connectionRadius: number;
}

interface GravityPoint {
  x: number;
  y: number;
  strength: number;
  radius: number;
}

const COLORS = ['#22D3EE', '#A855F7', '#EC4899', '#22C55E', '#F59E0B'];

// Static gravity points (percentage-based, calculated once)
const GRAVITY_POINTS_CONFIG = [
  { xPct: 0.2, yPct: 0.3, strength: 0.3, radius: 150 },
  { xPct: 0.8, yPct: 0.5, strength: 0.25, radius: 120 },
  { xPct: 0.5, yPct: 0.7, strength: 0.2, radius: 100 },
];

export function Particles() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationRef = useRef<number | undefined>(undefined);
  const particlesRef = useRef<Particle[]>([]);
  const mouseRef = useRef<{ x: number; y: number }>({ x: -1000, y: -1000 });
  const gravityPointsRef = useRef<GravityPoint[]>([]);
  const timeRef = useRef<number>(0);
  const scrollVelocityRef = useRef<number>(0);
  const lastScrollYRef = useRef<number>(0);
  const scrollTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const centerXRef = useRef<number>(0);
  const centerYRef = useRef<number>(0);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const resizeCanvas = () => {
      canvas.width = canvas.offsetWidth;
      canvas.height = canvas.offsetHeight;
      centerXRef.current = canvas.width / 2;
      centerYRef.current = canvas.height / 2;
    };

    const createParticle = (_: unknown, __?: number): Particle => {
      const type = Math.random();
      const baseSpeed = type < 0.15 ? 4 : type < 0.4 ? 3 : type < 0.7 ? 1 : 2;
      const vx = (Math.random() - 0.5) * baseSpeed;
      const vy = (Math.random() - 0.5) * baseSpeed;

      // Radius ranges from ~0.5 to ~5
      const radius = type < 0.15 ? Math.random() * 1.5 + 0.5 : type < 0.4 ? Math.random() * 2.5 + 1.5 : type < 0.7 ? Math.random() * 3 + 2 : Math.random() * 2 + 1;

      // Bigger particles are brighter: alpha scales with radius (0.1 to 0.7)
      const alpha = 0.1 + (radius / 5) * 0.6;

      // Random connection radius for each particle (80-250px)
      const connectionRadius = Math.random() * 170 + 80;

      return {
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        vx,
        vy,
        baseVx: vx,
        baseVy: vy,
        radius,
        baseRadius: radius,
        color: COLORS[Math.floor(Math.random() * COLORS.length)],
        alpha,
        trail: [],
        waveOffset: Math.random() * Math.PI * 2,
        waveAmplitude: Math.random() * 0.5 + 0.3,
        connectionRadius,
      };
    };

    const initParticles = () => {
      const particleCount = Math.floor((canvas.width * canvas.height) / 5000);
      particlesRef.current = Array.from({ length: particleCount }, createParticle);
    };

    const initGravityPoints = () => {
      // Static gravity points - only positions update on resize
      gravityPointsRef.current = GRAVITY_POINTS_CONFIG.map((gp) => ({
        x: canvas.width * gp.xPct,
        y: canvas.height * gp.yPct,
        strength: gp.strength,
        radius: gp.radius,
      }));
    };

    const drawTrail = (p: Particle) => {
      for (let i = 0; i < p.trail.length; i++) {
        const t = p.trail[i];
        const trailRadius = p.radius * (i / p.trail.length) * 0.8;
        ctx.beginPath();
        ctx.arc(t.x, t.y, trailRadius, 0, Math.PI * 2);
        ctx.fillStyle = p.color;
        ctx.globalAlpha = t.alpha * 0.4;
        ctx.fill();
      }
      ctx.globalAlpha = 1;
    };

    const drawParticle = (p: Particle) => {
      // Draw trail first
      drawTrail(p);

      // Draw glow
      const gradient = ctx.createRadialGradient(p.x, p.y, 0, p.x, p.y, p.radius * 2);
      gradient.addColorStop(0, p.color);
      gradient.addColorStop(1, 'transparent');
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.radius * 2, 0, Math.PI * 2);
      ctx.fillStyle = gradient;
      ctx.globalAlpha = p.alpha * 0.3;
      ctx.fill();

      // Draw particle
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2);
      ctx.fillStyle = p.color;
      ctx.globalAlpha = p.alpha;
      ctx.fill();
      ctx.globalAlpha = 1;
    };

    const drawConnections = () => {
      const particles = particlesRef.current;

      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x;
          const dy = particles[i].y - particles[j].y;
          const distance = Math.sqrt(dx * dx + dy * dy);

          // Use the larger connectionRadius of either particle
          const maxDist = Math.max(particles[i].connectionRadius, particles[j].connectionRadius);

          if (distance < maxDist) {
            const alpha = (1 - distance / maxDist) * 0.25;
            ctx.beginPath();
            ctx.moveTo(particles[i].x, particles[i].y);
            ctx.lineTo(particles[j].x, particles[j].y);
            ctx.strokeStyle = particles[i].color;
            ctx.globalAlpha = alpha;
            ctx.lineWidth = 0.5;
            ctx.stroke();
            ctx.globalAlpha = 1;
          }
        }
      }
    };

    const updateParticle = (p: Particle) => {
      const scrollSpeed = Math.abs(scrollVelocityRef.current);

      if (scrollSpeed > 0.1) {
        // Update trail
        p.trail.unshift({ x: p.x, y: p.y, alpha: p.alpha });
        if (p.trail.length > 10) p.trail.pop();
        p.trail.forEach((t) => (t.alpha *= 0.8));

        // Calculate direction from center to particle (starfield direction)
        const dx = p.x - centerXRef.current;
        const dy = p.y - centerYRef.current;
        const distFromCenter = Math.sqrt(dx * dx + dy * dy);

        // Normalize direction (avoid division by zero)
        const dirX = distFromCenter > 0 ? dx / distFromCenter : 0;
        const dirY = distFromCenter > 0 ? dy / distFromCenter : 0;

        // Bigger stars move faster (radius affects speed)
        const baseSpeed = 0.5 + (p.radius / 5) * 3;
        const speed = baseSpeed * (scrollSpeed / 2);

        // Scroll down = outward, scroll up = inward
        const direction = scrollVelocityRef.current > 0 ? 1 : -1;

        // Apply starfield movement
        p.x += dirX * speed * direction;
        p.y += dirY * speed * direction;

        // When going off screen, wrap to opposite side
        const margin = 50;
        if (p.x < -margin) p.x = canvas.width + margin;
        if (p.x > canvas.width + margin) p.x = -margin;
        if (p.y < -margin) p.y = canvas.height + margin;
        if (p.y > canvas.height + margin) p.y = -margin;

        // Scale radius based on distance from center (35% at center, 100% at edges)
        const maxDist = Math.max(canvas.width, canvas.height) * 0.5;
        const distRatio = Math.min(distFromCenter / maxDist, 1);
        p.radius = p.baseRadius * (0.35 + distRatio * 0.65);

        // When scrolling up and star gets close to center, move to edge
        if (direction < 0 && distFromCenter < 20) {
          const angle = Math.random() * Math.PI * 2;
          const edgeDist = Math.max(canvas.width, canvas.height) * 0.6;
          p.x = centerXRef.current + Math.cos(angle) * edgeDist;
          p.y = centerYRef.current + Math.sin(angle) * edgeDist;
          p.trail = [];
        }
      } else {
        // Fade out trails when not scrolling
        p.trail.forEach((t) => (t.alpha *= 0.9));
        if (p.trail.length > 0 && p.trail[p.trail.length - 1].alpha < 0.01) {
          p.trail.pop();
        }
      }
    };

    const animate = () => {
      timeRef.current++;
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      // Gentle ease-out over ~0.7s (0.96^42 ≈ 0.18, feels more linear)
      scrollVelocityRef.current *= 0.96;

      particlesRef.current.forEach(updateParticle);
      drawConnections();
      particlesRef.current.forEach(drawParticle);

      animationRef.current = requestAnimationFrame(animate);
    };

    const handleScroll = () => {
      const currentScrollY = window.scrollY;
      const delta = currentScrollY - lastScrollYRef.current;
      // Add momentum - blend new scroll with existing velocity for smoother feel
      scrollVelocityRef.current = scrollVelocityRef.current * 0.5 + delta * 0.5;
      lastScrollYRef.current = currentScrollY;
    };

    const handleMouseMove = (e: MouseEvent) => {
      const rect = canvas.getBoundingClientRect();
      mouseRef.current.x = e.clientX - rect.left;
      mouseRef.current.y = e.clientY - rect.top;
    };

    const handleMouseLeave = () => {
      mouseRef.current.x = -1000;
      mouseRef.current.y = -1000;
    };

    resizeCanvas();
    initParticles();
    initGravityPoints();
    lastScrollYRef.current = window.scrollY;
    animate();

    const handleResize = () => {
      resizeCanvas();
      initParticles();
      initGravityPoints();
    };

    window.addEventListener('resize', handleResize);
    window.addEventListener('scroll', handleScroll, { passive: true });
    canvas.addEventListener('mousemove', handleMouseMove);
    canvas.addEventListener('mouseleave', handleMouseLeave);

    return () => {
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('scroll', handleScroll);
      canvas.removeEventListener('mousemove', handleMouseMove);
      canvas.removeEventListener('mouseleave', handleMouseLeave);
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
      if (scrollTimeoutRef.current) {
        clearTimeout(scrollTimeoutRef.current);
      }
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="absolute inset-0 w-full h-full"
      style={{ opacity: 0.85 }}
    />
  );
}
