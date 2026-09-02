# 09: Import/Export Dialogs

**What to build:** Import and Export modal dialogs for bringing in external API definitions and sharing work.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

### Import Dialog
- [ ] Modal opens when Import is clicked in TopBar
- [ ] Format selection: OpenAPI/Swagger, Postman Collection, Insomnia, cURL, HAR, Reqly
- [ ] File picker: drag-and-drop zone or browse button
- [ ] Destination selector: workspace dropdown
- [ ] Preview shows: Collections found, Requests found, Environments found, Variables found, Conflicts, Warnings
- [ ] Conflict resolution: Skip (default), Merge, Overwrite per conflicting item
- [ ] Import button processes the file
- [ ] Cancel button closes the dialog
- [ ] Progress indicator during import
- [ ] Success/error notification after import

### Export Dialog
- [ ] Modal opens when Export is clicked in TopBar
- [ ] Format selection: Collection, Workspace, OpenAPI, cURL, HAR, Environment, Documentation
- [ ] Options: Include secrets, Include tests, Include scripts, Include docs, Normalize variables
- [ ] Export button generates the file
- [ ] Cancel button closes the dialog
- [ ] Download the exported file

- [ ] All components use theme tokens
