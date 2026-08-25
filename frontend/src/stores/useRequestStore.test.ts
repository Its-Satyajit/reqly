import { describe, expect, it } from 'vitest'
import { emptyTabDraft, fileInputFromDraft } from './useRequestStore'

describe('fileInputFromDraft request settings', () => {
  it('defaults to no explicit settings', () => {
    const input = fileInputFromDraft(emptyTabDraft())
    expect(input.timeout).toBeUndefined()
    expect(input.followRedirects).toBeUndefined()
  })

  it('carries timeout and redirect overrides into the save shape', () => {
    const input = fileInputFromDraft({
      ...emptyTabDraft(),
      timeout: 5000,
      followRedirects: false,
    })
    expect(input.timeout).toBe(5000)
    expect(input.followRedirects).toBe(false)
  })
})
