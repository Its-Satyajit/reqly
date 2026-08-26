import { Toast } from "@base-ui/react/toast"

export type ToastType = "success" | "error" | "warning" | "info"

export interface ToastOptions {
	description?: React.ReactNode
	timeout?: number
}

const manager = Toast.createToastManager<{ title: React.ReactNode; description?: React.ReactNode; type: ToastType }>()

function add(type: ToastType, title: string, options?: ToastOptions): string {
	return manager.add({
		title,
		description: options?.description,
		type,
		timeout: options?.timeout,
	})
}

export const toast = Object.assign(manager, {
	success: (title: string, options?: ToastOptions) => add("success", title, options),
	error: (title: string, options?: ToastOptions) => add("error", title, options),
	warning: (title: string, options?: ToastOptions) => add("warning", title, options),
	info: (title: string, options?: ToastOptions) => add("info", title, options),
})
