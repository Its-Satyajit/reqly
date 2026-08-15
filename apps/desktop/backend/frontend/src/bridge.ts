import { AppService } from '../bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/index'
import type {
  RequestInput,
  RequestSender,
  ResponseData,
} from '@reqly/frontend'
import { useRequestStore } from '@reqly/frontend'

/**
 * wailsSender executes requests through the Go core via the generated Wails
 * bindings, then normalizes the core response into the shared ResponseData
 * shape the UI renders.
 */
export const wailsSender: RequestSender = async (req: RequestInput): Promise<ResponseData> => {
  const res = await AppService.SendRequest(req as never)
  if (!res) {
    throw new Error('core returned an empty response')
  }
  return res as ResponseData
}

/**
 * Wires the Go core behind the shared request store. Called once from the
 * host entry point, before the React tree mounts.
 */
export function initRequestBridge(): void {
  useRequestStore.getState().setSender(wailsSender)
}
