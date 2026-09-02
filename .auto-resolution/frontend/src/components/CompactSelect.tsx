import { cn } from "#lib/utils";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#components/ui/select";

export interface SelectOption {
	value: string;
	label: string;
}

interface CompactSelectProps {
	value: string;
	onChange: (value: string) => void;
	options: SelectOption[];
	ariaLabel: string;
	className?: string;
	disabled?: boolean;
}

/** The app's single dropdown device — Base UI Select styled for the dense
 * chrome. Replaces every native <select> so menus match the theme. */
export function CompactSelect({
	value,
	onChange,
	options,
	ariaLabel,
	className,
	disabled,
}: CompactSelectProps) {
	return (
		<Select
			items={options}
			value={value}
			onValueChange={(next) => {
				if (next !== null) onChange(next);
			}}
			disabled={disabled}
		>
			<SelectTrigger
				aria-label={ariaLabel}
				className={cn("h-7 w-auto gap-1 rounded-md px-2 text-xs", className)}
			>
				<SelectValue />
			</SelectTrigger>
			<SelectContent className="max-h-72 min-w-(--anchor-width)">
				{options.map((option) => (
					<SelectItem
						key={option.value}
						value={option.value}
						className="text-xs"
					>
						{option.label}
					</SelectItem>
				))}
			</SelectContent>
		</Select>
	);
}
