import { Toast as ToastPrimitive } from "@base-ui/react/toast"

/** The app-wide toast manager instance. Kept out of toast.tsx so that file
 * only exports components. */
export const toast = ToastPrimitive.createToastManager()
