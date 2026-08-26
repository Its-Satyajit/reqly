import { Input as InputPrimitive } from "@base-ui/react/input"

import { cn } from "#lib/utils"

function Input({ className, ...props }: InputPrimitive.Props) {
	return (
		<InputPrimitive
			data-slot="input"
			className={cn(
				"h-8 w-full min-w-0 rounded-md border border-input bg-background px-2.5 py-1 text-xs text-foreground shadow-xs outline-none transition-[color,box-shadow]",
				"placeholder:text-muted-foreground disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50",
				"focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50",
				"aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20",
				className,
			)}
			{...props}
		/>
	)
}

export { Input }