import { cn } from '@/lib/utils'

interface LinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  href: string
  children: React.ReactNode
  external?: boolean
}

export function Link({ href, children, external, className, ...props }: LinkProps) {
  return (
    <a
      href={href}
      target={external ? '_blank' : undefined}
      rel={external ? 'noopener noreferrer' : undefined}
      className={cn(className)}
      {...props}
    >
      {children}
    </a>
  )
}
