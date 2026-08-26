import { toast } from "sonner";

export function notifySuccess(title: string, description?: string): void {
	toast.success(title, { description });
}

export function notifyError(title: string, description?: string): void {
	toast.error(title, { description });
}

export function notifyWarning(title: string, description?: string): void {
	toast.warning(title, { description });
}
