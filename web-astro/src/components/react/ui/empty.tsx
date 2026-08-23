import * as React from 'react'
import { Inbox } from 'lucide-react'
import { cn } from '@/lib/utils'

interface EmptyProps extends React.HTMLAttributes<HTMLDivElement> {
  title?: string
  description?: string
}

function Empty({ title = '暂无数据', description, className, children, ...props }: EmptyProps) {
  return (
    <div
      className={cn('flex flex-col items-center justify-center py-16 text-center', className)}
      {...props}
    >
      <div className="mb-4">
        <Inbox className="h-12 w-12 text-muted-foreground" />
      </div>
      <h3 className="text-lg font-medium text-foreground">{title}</h3>
      {description && <p className="mt-2 max-w-md text-sm text-muted-foreground">{description}</p>}
      {children && <div className="mt-6">{children}</div>}
    </div>
  )
}

export { Empty }
