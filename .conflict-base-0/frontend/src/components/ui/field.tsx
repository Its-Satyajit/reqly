import { Field } from "@base-ui/react/field"

import { cn } from "#lib/utils"

/** Label wraps Base UI's Field.Label, matching the shadcn/base-ui label
 * styling used across the Reqly editor surfaces. */
function Label({ className, ...props }: Field.Label.Props) {
	return (
		<Field.Label
			data-slot="label"
			className={cn(
				"text-xs font-medium text-muted-foreground select-none",
				className,
			)}
			{...props}
		/>
	)
}

export { Label }