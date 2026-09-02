# 04: Collections Explorer

**What to build:** The Collections tree in the Context Sidebar for navigating API structure. Supports drag-and-drop reordering and context menu actions.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Collections tree renders in Context Sidebar when activeView is 'requests'
- [ ] Tree shows collections, folders, and requests with appropriate icons
- [ ] Tree supports expand/collapse for collections and folders
- [ ] Tree supports search/filter
- [ ] Drag-and-drop reordering works for collections, folders, and requests
- [ ] Drag-and-drop updates the underlying collection file
- [ ] Context menu shows: Rename, Move, Duplicate, Delete, Run, Import, Export, Generate Docs, Generate Tests, Generate Mock
- [ ] Context menu actions trigger the correct operation
- [ ] New Collection, New Folder, New Request buttons at bottom of tree
- [ ] Tree uses theme tokens for all colors and icons
- [ ] Tree supports keyboard navigation (arrow keys, Enter to select)
